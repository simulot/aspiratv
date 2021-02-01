package artetv

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/simulot/aspiratv/metadata/nfo"
	"github.com/simulot/aspiratv/providers"
	"github.com/simulot/aspiratv/providers/matcher"

	"github.com/simulot/aspiratv/net/myhttp/httptest"

	"github.com/gocolly/colly"

	"github.com/simulot/aspiratv/net/myhttp"
	"github.com/simulot/aspiratv/parsers/htmlparser"
)

// init registers ArteTV provider
func init() {
	p, err := New()
	if err != nil {
		panic(err)
	}
	providers.Register(p)
}

// Provider constants
const (
	arteIndex      = "https://www.arte.tv"
	arteCDN        = "https://static-cdn.arte.tv"
	arteCollection = "https://www.arte.tv/guide/api/api/zones/fr/collection_videos/?id=%s&page=%d"         // Id and Page
	arteSearch     = "https://www.arte.tv/guide/api/api/zones/fr/listing_SEARCH/?page=1&limit=20&query=%s" // Search term
)

// Track when this is the first time the Show is invoked
var runCounter = 0

type getter interface {
	Get(ctx context.Context, uri string) (io.ReadCloser, error)
}

// ArteTV structure handles arte  catalog of shows
type ArteTV struct {
	config            providers.Config
	getter            getter
	preferredVersions []string // versionCode List of version in order of preference VF,VA...
	preferredQuality  []string
	preferredMedia    string // mediaType mp4,hls
	htmlParserFactory *htmlparser.Factory
	seenPrograms      map[string]bool
	deadline          time.Duration
}

// WithGetter inject a getter in FranceTV object instead of normal one
func WithGetter(g getter) func(p *ArteTV) {
	return func(p *ArteTV) {
		p.getter = g

	}
}

type throttler struct {
	g        getter
	ticker   *time.Ticker
	throttle chan struct{}
	burst    int
	rate     time.Duration
	stop     chan struct{}
	once     sync.Once
}

func newThrottler(g getter, rate int, burst int) *throttler {
	// lazy initialisation
	return &throttler{
		g:     g,                                 // The orriginal getter
		burst: burst,                             // allow a burst of queries
		rate:  time.Second / time.Duration(rate), // Query
		stop:  make(chan struct{}),               // Stop me if you can
	}
}

func (t *throttler) init() {
	t.throttle = make(chan struct{}, t.burst)
	t.ticker = time.NewTicker(t.rate)
	for i := 0; i < t.burst; i++ {
		t.throttle <- struct{}{}
	}
	go func() {
		defer t.ticker.Stop()
		for {
			select {
			case <-t.stop:
				return
			case <-t.ticker.C:
				t.throttle <- struct{}{}
			default:
			}
		}

	}()
}

func (t *throttler) Stop() {
	<-t.stop
}

func (t *throttler) Get(ctx context.Context, uri string) (io.ReadCloser, error) {
	t.once.Do(t.init)
	<-t.throttle
	return t.g.Get(ctx, uri)
}

// New setup a Show provider for Arte
func New() (*ArteTV, error) {
	throttler := newThrottler(myhttp.DefaultClient, 2, 5)
	p := &ArteTV{
		getter: throttler,
		//TODO: get preferences from config file
		preferredVersions: []string{"VF", "VOF", "VF-STF", "VOF-STF", "VO-STF", "VOF-STMF", "VO"}, // "VF-STMF" "VA", "VA-STA"
		preferredMedia:    "mp4",
		preferredQuality:  []string{"SQ", "XQ", "EQ", "HQ", "MQ"},
		htmlParserFactory: htmlparser.NewFactory(),
		seenPrograms:      map[string]bool{},
		deadline:          30 * time.Second,
	}
	return p, nil
}

func (p *ArteTV) Configure(c providers.Config) {
	p.config = c
	if p.config.Log.IsDebug() {
		p.deadline = 1 * time.Hour
	}

}

// withGetter set a getter for ArteTV
func withGetter(g getter) func(p *ArteTV) {
	return func(p *ArteTV) {
		p.getter = g
	}
}

// Name return the name of the provider
func (p ArteTV) Name() string { return "artetv" }

// Shows download the shows catalog from the web site.
func (p *ArteTV) MediaList(ctx context.Context, mm []*matcher.MatchRequest) chan *providers.Media {
	shows := make(chan *providers.Media)

	go func() {
		defer close(shows)
		for _, m := range mm {
			if ctx.Err() != nil {
				return
			}
			if m.Provider == p.Name() {
				for s := range p.getShowList(ctx, m) {
					shows <- s
				}
			}
		}
	}()

	return shows
}

func (p *ArteTV) getShowList(ctx context.Context, mr *matcher.MatchRequest) chan *providers.Media {
	shows := make(chan *providers.Media)

	go func() {
		defer func() {
			close(shows)

		}()

		matchedSeries := []Data{}
		matchedShows := []Data{}
		page := 1

		//TODO: use user's preferred language
		const apiSEARCH = "https://www.arte.tv/guide/api/emac/v3/fr/web/data/SEARCH_LISTING"

		u, err := url.Parse(apiSEARCH)
		if err != nil {
			p.config.Log.Error().Printf("[%s] Can't call search API: %q", p.Name(), err)
			return
		}
		v := u.Query()
		// v.Set("imageFormats", "square,banner,landscape,poster")
		v.Set("imageFormats", "*")
		v.Set("query", mr.Show)
		v.Set("mainZonePage", "1")
		v.Set("page", strconv.Itoa(page))
		v.Set("limit", "100")
		for {
			u.RawQuery = v.Encode()

			p.config.Log.Debug().Printf("[%s] Search url: %q", p.Name(), u.String())
			var result APIResult
			ctxLocal, doneLocal := context.WithTimeout(ctx, p.deadline)

			r, err := p.getter.Get(ctxLocal, u.String())
			if err != nil {
				p.config.Log.Error().Printf("[%s] Can't call search API: %q", p.Name(), err)
				doneLocal()
				return
			}

			defer r.Close()

			if p.config.Log.IsDebug() {
				r = httptest.DumpReaderToFile(p.config.Log, r, "artetv-search-")
			}

			err = json.NewDecoder(r).Decode(&result)
			if err != nil {
				p.config.Log.Error().Printf("[%s] Can't decode search API result: %q", p.Name(), err)
				doneLocal()
				return
			}
			if ctxLocal.Err() != nil {
				p.config.Log.Error().Printf("%s", ctxLocal.Err())
				doneLocal()
				return
			}

			doneLocal()

			for _, d := range result.Data {
				if strings.Contains(strings.ToLower(d.Title), mr.Show) {
					if d.Kind.IsCollection {
						matchedSeries = append(matchedSeries, d)
					}
					if !p.config.KeepBonus && (d.Kind.Code != "SHOW") {
						continue
					}
					if strings.ToLower(d.Title) == mr.Show {
						matchedShows = append(matchedShows, d)
					}
				}
			}

			if len(result.NextPage) == 0 {
				break
			}
			page++
			v.Set("page", strconv.Itoa(page))

		}
		if len(matchedSeries) > 0 {
			for _, d := range matchedSeries {
				for info := range p.getSerie(ctx, mr, d) {
					shows <- info
				}
			}
			return
		}
		if len(matchedShows) > 0 {
			for s := range p.getShows(ctx, mr, matchedShows) {
				s.Match = mr
				shows <- s
			}
		}
	}()
	return shows
}

func (p *ArteTV) getShows(ctx context.Context, mr *matcher.MatchRequest, data []Data) chan *providers.Media {
	shows := make(chan *providers.Media)
	go func() {
		defer close(shows)
		// Emit media found in the current collection/season
		for _, ep := range data {
			media := &providers.Media{
				ID:       ep.ProgramID,
				ShowType: providers.Movie,
				Match:    mr,
			}

			info := nfo.Movie{
				MediaInfo: nfo.MediaInfo{
					UniqueID: []nfo.ID{
						{
							ID:   ep.ProgramID,
							Type: "ARTETV",
						},
					},
					Title: ep.Title,
					Plot:  ep.ShortDescription,
					Thumb: getThumbs(ep.Images),
					// TVShow:    &tvshow,
					Tag: []string{"Arte"},
				},
			}

			if len(ep.Subtitle) > 0 {
				info.MediaInfo.Title += " - " + ep.Subtitle
			}

			// TODO Actors
			media.SetMetaData(&info)
			shows <- media
		}
	}()

	return shows
}

//https://www.arte.tv/guide/api/emac/v3/fr/web/programs/044892-008-A/?
//https://    www.arte.tv/guide/api/emac/v3/fr/web/data/COLLECTION_VIDEOS/?collectionId=RC-014408&page=1&limit=100
//https://api-cdn.arte.tv/      api/emac/v3/fr/web/data/COLLECTION_VIDEOS/?collectionId=RC-015842&page=2&limit=12
var (
	parseSeason  = regexp.MustCompile(`Saison (\d+)`)              // Detect season number in web page
	parseEpisode = regexp.MustCompile(`^(?:.+) \((\d+)\/(\d+)\)$`) // Extract episode number
)

// getSerie
// Arte presents a serie either as collection of episodes for a single season or as a collection of collection of episodes for multiple seasons.

var parseShowSeason = regexp.MustCompile(`^(.+) - Saison (\d+)$`)

func (p *ArteTV) getSerie(ctx context.Context, mr *matcher.MatchRequest, d Data) chan *providers.Media {
	ctx, done := context.WithTimeout(ctx, p.deadline)
	shows := make(chan *providers.Media)

	go func() {
		defer func() {
			close(shows)
			done()
		}()

		// TODO: use user's preferred language
		// Refactoring. The collection's page contains a script with the full serie split per seasons, no need to call the api

		parser := p.htmlParserFactory.New()
		pgm := InitialProgram{}

		parser.OnHTML("body > script", func(e *colly.HTMLElement) {
			if strings.Index(e.Text, "__INITIAL_STATE__") < 0 {
				return
			}
			// Get JSON with collection data from the HTML page of the collection
			start := strings.Index(e.Text, "{")
			end := strings.LastIndex(e.Text, "}")
			if start < 0 || end < 0 {
				return
			}

			js := e.Text[:end+1][start:]
			err := json.NewDecoder(strings.NewReader(js)).Decode(&pgm)
			if err != nil {
				p.config.Log.Error().Printf("[%s] Can't parse JSON collection: %w", p.Name(), err)
				return
			}

		})

		err := parser.Visit(d.URL)
		if err != nil {
			p.config.Log.Error().Printf("[%s] Can't visit URL: %w", p.Name(), err)
			return
		}

		for _, page := range pgm.Pages.List {
			var tvshow nfo.TVShow
			for _, zone := range page.Zones {

				// Get show level info
				if zone.Code.Name == "collection_content" {
					for _, data := range zone.Data {
						tvshow.Title = data.Title
						tvshow.Plot = data.Description
						tvshow.Thumb = getThumbs(data.Images)
					}
					continue
				}

				// Get episodes
				if zone.Code.Name == "collection_subcollection" || zone.Code.Name == "collection_videos" {
					season := 0

					m := parseShowSeason.FindAllStringSubmatch(zone.Title, -1)
					if len(m) > 0 {
						season, _ = strconv.Atoi(m[0][2])
					}

					for _, ep := range zone.Data {
						episode := 0
						m := parseEpisode.FindAllStringSubmatch(ep.Title, -1)
						if len(m) > 0 {
							episode, _ = strconv.Atoi(m[0][1])
						}

						// Emit media found in the current collection/season
						media := &providers.Media{
							ID:       ep.ProgramID,
							ShowType: providers.Series,
							Match:    mr,
						}

						info := nfo.EpisodeDetails{
							MediaInfo: nfo.MediaInfo{
								UniqueID: []nfo.ID{
									{
										ID:   ep.ProgramID,
										Type: "ARTETV",
									},
								},
								Title:     getFirstString([]string{ep.Subtitle, ep.Title, tvshow.Title}),
								Showtitle: tvshow.Title,
								Plot:      getFirstString([]string{ep.Description, ep.ShortDescription}),
								Thumb:     getThumbs(ep.Images),
								TVShow:    &tvshow,
								Tag:       []string{"Arte"},
								Season:    season,
								Episode:   episode,
								Aired:     nfo.Aired(ep.Availability.Start),
								URL:       ep.URL,
							},
						}

						tvshow.HasEpisodes = len(zone.Data) > 0

						if ep.Kind.Code == "SHOW" && info.Season == 0 {
							info.Season = info.Aired.Time().Year()
						}

						if ep.Kind.Code == "BONUS" {
							info.Season = 0 // Specials
						}
						// TODO Actors --> details in player

						media.SetMetaData(&info)
						shows <- media
					}
				}
			}

		}
	}()

	return shows
}

func getFirstString(ss []string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}

func getThumbs(images map[string]Image) []nfo.Thumb {
	thumbs := []nfo.Thumb{}
	for k, i := range images {
		if len(i.BlurURL) == 0 {
			continue
		}
		aspect := "thumb"

		switch k {
		case "landscape":
			aspect = "thumb"
		case "banner":
			aspect = "fanart"
		case "portrait":
			aspect = "poster"
		case "square":
			aspect = "poster"
		}

		thumbs = append(thumbs, nfo.Thumb{
			Aspect:  aspect,
			Preview: i.BlurURL,
			URL:     getBestImage(i),
		})

	}
	return thumbs
}

// getBestImage retreive the url for the image of type "protrait/banner/landscape..." with the highest resolution
func getBestImage(image Image) string {
	bestResolution := 0
	bestURL := ""
	for _, r := range image.Resolutions {
		_ = 1
		res := r.H * r.W
		if res > bestResolution {
			bestURL = r.URL
			bestResolution = res
		}
	}
	if bestResolution == 0 {
		return ""
	}
	return bestURL
}

// https://api.arte.tv/api/player/v1/config/fr/083668-012-A?autostart=1&lifeCycle=1

const arteDetails = "https://api.arte.tv/api/player/v1/config/fr/%s?autostart=1&lifeCycle=1" // Player to get Video streams ProgID

// GetMediaDetails return the show's URL, a mp4 file
func (p *ArteTV) GetMediaDetails(ctx context.Context, m *providers.Media) error {
	info := m.Metadata.GetMediaInfo()

	player_url := fmt.Sprintf(arteDetails, info.UniqueID[0].ID) // TODO search for ARTE ID

	if !info.IsDetailed {
		parser := p.htmlParserFactory.New()
		pgm := InitialProgram{}
		js := ""

		parser.OnHTML("body > script", func(e *colly.HTMLElement) {
			if strings.Index(e.Text, "__INITIAL_STATE__") < 0 {
				return
			}
			// Get JSON with collection data from the HTML page of the collection
			start := strings.Index(e.Text, "{")
			end := strings.LastIndex(e.Text, "}")
			if start < 0 || end < 0 {
				return
			}

			js = e.Text[:end+1][start:]

		})

		err := parser.Visit(info.URL)
		if err != nil {
			p.config.Log.Error().Printf("[%s] Can't visit URL: %w", p.Name(), err)
			return err
		}
		err = json.NewDecoder(strings.NewReader(js)).Decode(&pgm)
		if err != nil {
			p.config.Log.Error().Printf("[%s] Can't parse JSON collection: %w", p.Name(), err)
			return err
		}
		for _, page := range pgm.Pages.List {
			for _, zone := range page.Zones {
				// Get episodes details
				if zone.Code.Name == "program_content" {
					for _, ep := range zone.Data {
						for _, credit := range ep.Credits {
							switch credit.Code {
							case "ACT":
								regActors := regexp.MustCompile(`^(.+)(?:\s\((.+)\))$|(.+)$`)

								for _, v := range credit.Values {
									actor := nfo.Actor{}

									m := regActors.FindAllStringSubmatch(v, -1)
									if len(m) > 0 {
										if len(m[0]) == 4 {
											if len(m[0][3]) > 0 {
												actor.Name = m[0][3]
											} else {
												actor.Name = m[0][1]
												actor.Role = m[0][2]
											}
										}
									}
									info.Actor = append(info.Actor, actor)
								}
							case "REA":
								for _, v := range credit.Values {
									info.Director = append(info.Director, v)
								}
							case "COUNTRY", "PRODUCTION_YEAR":
								for _, v := range credit.Values {
									info.Tag = append(info.Tag, v)
								}
							default:
								for _, v := range credit.Values {
									info.Credits = append(info.Credits, fmt.Sprintf("%s (%s)", v, credit.Label))
								}
							}
						}
					}
				}
			}
		}
	}

	p.config.Log.Trace().Printf("[%s] Title '%s' player url: %q", p.Name(), info.Title, player_url)

	r, err := p.getter.Get(ctx, player_url)
	if err != nil {
		return fmt.Errorf("Can't get player for %q: %w", info.Title, err)
	}
	if p.config.Log.IsDebug() {
		r = httptest.DumpReaderToFile(p.config.Log, r, "artetv-info-"+m.ID+"-")
	}
	defer r.Close()
	player := playerAPI{}
	err = json.NewDecoder(r).Decode(&player)
	if err != nil {
		return fmt.Errorf("Can't decode show's detailled information: %w", err)
	}

	u, err := p.getBestVideo(player.VideoJSONPlayer.VSR)
	if err != nil {
		return err
	}

	info.URL = u
	if info.Aired.Time().IsZero() {
		info.Aired = nfo.Aired(player.VideoJSONPlayer.VRA.Time())
	}

	// if info.TVShow != nil && !info.TVShow.HasEpisodes && info.Episode == 0 {
	// 	info.Season = info.Aired.Time().Year()
	// }
	return nil
}

type mapStrInt map[string]uint64

// getBestVideo return the best video stream given preferences
//   Streams are scored in following order:
//   - Stream quality, the highest possible
//   - Version (VF,VF_ST) that match preference

func (p *ArteTV) getBestVideo(ss map[string]StreamInfo) (string, error) {
	for _, v := range p.preferredVersions {
		for _, r := range p.preferredQuality {
			for _, s := range ss {
				if s.Quality == r && s.VersionCode == v {
					return s.URL, nil
				}
			}
		}
	}
	return "", errors.New("Can't find a suitable video stream")
}
