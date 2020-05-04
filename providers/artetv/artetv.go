package artetv

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/simulot/aspiratv/metadata/nfo"
	"github.com/simulot/aspiratv/providers"

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
	throttler := newThrottler(myhttp.DefaultClient, 20, 25)
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
func (p *ArteTV) MediaList(ctx context.Context, mm []*providers.MatchRequest) chan *providers.Media {
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

func (p *ArteTV) getShowList(ctx context.Context, mr *providers.MatchRequest) chan *providers.Media {
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

func (p *ArteTV) getShows(ctx context.Context, mr *providers.MatchRequest, data []Data) chan *providers.Media {
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
	parseCollectionInURL = regexp.MustCompile(`RC-\d+`)       // Detect Season URL
	parseSeason          = regexp.MustCompile(`Saison (\d+)`) // Detect season number in web page
)

// getSerie
// Arte presents a serie either as collection of episodes for a single season or as a collection of collection of episodes for multiple seasons.
func (p *ArteTV) getSerie(ctx context.Context, mr *providers.MatchRequest, d Data) chan *providers.Media {
	ctx, done := context.WithTimeout(ctx, p.deadline)
	shows := make(chan *providers.Media)

	go func() {
		defer func() {
			close(shows)
			done()
		}()

		//TODO: use user's preferred language
		const apiSEARCH = "https://www.arte.tv/guide/api/emac/v3/fr/web/data/COLLECTION_VIDEOS/?collectionId=%s&page=%d&limit=12"

		collectionIDs := map[string]string{"": d.ProgramID} // Collection per season
		seasonSearched := false

		tvshow := nfo.TVShow{
			Title: d.Title,
			Plot:  d.ShortDescription,
			Thumb: getThumbs(d.Images),
		}

	collectionLoop:
		for len(collectionIDs) > 0 {
			if ctx.Err() != nil {
				return
			}

			seasons := []string{}
			for k := range collectionIDs {
				seasons = append(seasons, k)
			}
			sort.Strings(seasons)
			u := fmt.Sprintf(apiSEARCH, collectionIDs[seasons[0]], 1)

			// Loop collections's pages
			for len(u) > 0 {

				u2, err := url.Parse(u)
				if err != nil {
					p.config.Log.Error().Printf("[%s] GetSeries: '%s' ", p.Name(), err)
					return
				}
				u2.Host = "www.arte.tv"
				u2.Path = "guide/api/emac/v3/fr/web/data/COLLECTION_VIDEOS"
				u = u2.String()
				p.config.Log.Trace().Printf("[%s] GetSeries URL is '%s'", p.Name(), u)
				r, err := p.getter.Get(ctx, u)

				if p.config.Log.IsDebug() {
					r = httptest.DumpReaderToFile(p.config.Log, r, "artetv-getcollection-")
				}
				if err != nil {
					p.config.Log.Error().Printf("[%s] Can't get collection: %q", p.Name(), err)
					return
				}
				if ctx.Err() != nil {
					return
				}
				var result APIResult
				err = json.NewDecoder(r).Decode(&result)
				r.Close()
				if err != nil {
					p.config.Log.Error().Printf("[%s] Can't get decode collection: %q", p.Name(), err)
					return
				}

				if len(result.Data) == 0 {
					// A collection of collection (a series, indeed) entry hasn't any Data. We have to fetch collections for each season
					if seasonSearched {
						p.config.Log.Error().Printf("[%s] Can't found collection with ID(%s): %q", p.Name(), d.ProgramID, err)
						return
					}
					seasonSearched = true

					// No results on a collection ID? this means this is a collection of collections...
					// Let's scrap the web page to get the collection list, most likely all seasons
					delete(collectionIDs, "")

					parser := p.htmlParserFactory.New()

					parser.OnHTML("a.next-navbar__slide", func(e *colly.HTMLElement) {
						var season, id string
						m := parseSeason.FindAllStringSubmatch(e.Text, -1)
						if len(m) == 1 {
							season = m[0][1]
						}
						m = parseCollectionInURL.FindAllStringSubmatch(e.Attr("href"), -1)
						if len(m) == 2 {
							id = m[1][0]
						}
						collectionIDs[season] = id
					})

					err := parser.Visit(d.URL)
					if err != nil {
						p.config.Log.Error().Printf("[%s] Can't get collection: %q", p.Name(), err)
						return
					}
					continue collectionLoop
				}

				// Emit media found in the current collection/season
				for _, ep := range result.Data {
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
							Title:     ep.Title,
							Showtitle: tvshow.Title,
							Plot:      ep.ShortDescription,
							Thumb:     getThumbs(ep.Images),
							TVShow:    &tvshow,
							Tag:       []string{"Arte"},
						},
					}
					setEpisodeFormTitle(&info, tvshow, ep)
					if info.Episode != 0 && info.Season != 0 {
						tvshow.HasEpisodes = true
					}

					if ep.Kind.Code == "BONUS" {
						info.Season = 0 // Specials
					}
					if tvshow.HasEpisodes && info.Episode == 0 {
						info.Season = 0 // Specials
					}
					if !tvshow.HasEpisodes && info.Episode == 0 {
						info.Season = info.Aired.Time().Year()
					}

					// TODO Actors

					media.SetMetaData(&info)
					shows <- media
				}
				u = result.NextPage
			}
			delete(collectionIDs, seasons[0]) // Season on top of the stack is done.
		}
	}()

	return shows
}

var (
	parseTitleSeasonEpisode = regexp.MustCompile(`^(.+) - Saison (\d+) \((\d+)\/\d+\)$`)
	parseTitleEpisode       = regexp.MustCompile(`^(.+) \((\d+)\/\d+\)$`)
	removeHiphens           = regexp.MustCompile(`(?m)-{2,}|^\s*-*\s|\s*-\s*$`)
	removeSpaces            = regexp.MustCompile(`(?m)^\s+|\s+$|\s{2,}`)
)

// setEpisodeFormTitle tries to normalize episode titles despite coding inconsistencies
func setEpisodeFormTitle(show *nfo.EpisodeDetails, tvshow nfo.TVShow, ep Data) {

	show.Showtitle = tvshow.Title

	m := parseTitleSeasonEpisode.FindAllStringSubmatch(ep.Title, -1)
	if len(m) > 0 {
		ep.Title = m[0][1]
		show.Season, _ = strconv.Atoi(m[0][2])
		show.Episode, _ = strconv.Atoi(m[0][3])
	}
	m = parseTitleEpisode.FindAllStringSubmatch(ep.Title, -1)
	if len(m) > 0 {
		ep.Title = m[0][1]
		show.Season = 1
		show.Episode, _ = strconv.Atoi(m[0][2])
	}

	title := ep.Title + " " + ep.Subtitle
	title = strings.ReplaceAll(title, tvshow.Title, " ")
	title = removeHiphens.ReplaceAllLiteralString(title, "")
	title = removeSpaces.ReplaceAllLiteralString(title, "")

	show.Title = title
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

	if info.URL != "" {
		return nil
	}

	url := fmt.Sprintf(arteDetails, m.ID)
	p.config.Log.Trace().Printf("[%s] Title '%s' url: %q", p.Name(), m.Metadata.GetMediaInfo().Title, url)
	r, err := p.getter.Get(ctx, url)
	if err != nil {
		return fmt.Errorf("Can't get show's detailled information: %w", err)
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
	info.Aired = nfo.Aired(player.VideoJSONPlayer.VRA.Time())

	if info.TVShow != nil && !info.TVShow.HasEpisodes && info.Episode == 0 {
		info.Season = info.Aired.Time().Year()
	}
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
