package artetv

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"

	"github.com/simulot/aspiratv/net/http"
	"github.com/simulot/aspiratv/parsers/htmlparser"
	"github.com/simulot/aspiratv/providers"
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
	Get(uri string) (io.Reader, error)
}

// ArteTV structure handles arte  catalog of shows
type ArteTV struct {
	getter            getter
	preferredVersions []string // versionCode List of version in order of preference VF,VA...
	preferredMedia    string   // mediaType mp4,hls
	debug             bool
	htmlParserFactory *htmlparser.Factory
	seenPrograms      map[string]bool
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

func (t *throttler) Get(uri string) (io.Reader, error) {
	t.once.Do(t.init)
	<-t.throttle
	return t.g.Get(uri)
}

// New setup a Show provider for Arte
func New(conf ...func(p *ArteTV)) (*ArteTV, error) {
	throttler := newThrottler(http.DefaultClient, 2, 100)
	p := &ArteTV{
		getter: throttler,
		//TODO: get preferences from config file
		preferredVersions: []string{"VF", "VOF", "VOF-STF", "VOSTF", "VF-STF"}, // "VF-STMF" "VA", "VA-STA"
		preferredMedia:    "mp4",
		htmlParserFactory: htmlparser.NewFactory(),
		seenPrograms:      map[string]bool{},
	}
	for _, fn := range conf {
		fn(p)
	}
	return p, nil
}

// SetDebug set debug mode
func (p *ArteTV) SetDebug(b bool) {
	p.debug = b
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
func (p *ArteTV) Shows(mm []*providers.MatchRequest) ([]*providers.Show, error) {
	shows := []*providers.Show{}

	for _, m := range mm {
		if m.Provider == p.Name() {
			r, err := p.getShowList(m)
			if err != nil {
				return nil, err
			}
			shows = append(shows, r...)
			for _, s := range shows {
				s.Destination = m.Destination
			}
		}
	}

	return shows, nil
}

func (p *ArteTV) getShowList(m *providers.MatchRequest) ([]*providers.Show, error) {
	//TODO: use user's preferred language
	const apiSEARCH = "https://www.arte.tv/guide/api/emac/v3/fr/web/data/SEARCH_LISTING"

	u, err := url.Parse(apiSEARCH)
	if err != nil {
		return nil, err
	}
	v := u.Query()
	v.Set("imageFormats", "square,banner,landscape")
	v.Set("query", m.Show)
	v.Set("mainZonePage", "1")
	v.Set("page", "1")
	v.Set("limit", "100")

	u.RawQuery = v.Encode()

	var result APIResult

	r, err := p.getter.Get(u.String())
	if err != nil {
		return nil, err
	}

	err = json.NewDecoder(r).Decode(&result)
	if err != nil {
		return nil, err
	}

	matchedSeries := []Data{}
	matchedShows := []Data{}

	for _, d := range result.Data {
		if strings.Contains(strings.ToLower(d.Title), m.Show) {
			if d.Kind.IsCollection {
				matchedSeries = append(matchedSeries, d)
			} else {
				matchedShows = append(matchedShows, d)
			}
		}
	}

	shows := []*providers.Show{}

	if len(matchedSeries) > 0 {
		for _, d := range matchedSeries {
			ss, err := p.getSerie(d)
			if err != nil {
				return nil, err
			}
			shows = append(shows, ss...)
		}
	}
	return shows, nil
}

//https://www.arte.tv/guide/api/emac/v3/fr/web/programs/044892-008-A/?
//https://    www.arte.tv/guide/api/emac/v3/fr/web/data/COLLECTION_VIDEOS/?collectionId=RC-014408&page=1&limit=100
//https://api-cdn.arte.tv/      api/emac/v3/fr/web/data/COLLECTION_VIDEOS/?collectionId=RC-015842&page=2&limit=12
var (
	parseCollectionInURL = regexp.MustCompile(`RC-\d+`)
	parseSeason          = regexp.MustCompile(`Saison (\d+)`)
)

func (p *ArteTV) getSerie(d Data) ([]*providers.Show, error) {
	//TODO: use user's preferred language
	const apiSEARCH = "https://www.arte.tv/guide/api/emac/v3/fr/web/data/COLLECTION_VIDEOS/?collectionId=%s&page=%d&limit=12"

	shows := []*providers.Show{}
	collectionIDs := map[string]string{"": d.ProgramID} // Collection per season

	seasonSearched := false

collectionLoop:
	for len(collectionIDs) > 0 {
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
				return nil, err
			}
			u2.Host = "www.arte.tv"
			u2.Path = "guide/api/emac/v3/fr/web/data/COLLECTION_VIDEOS"
			u = u2.String()

			if p.debug {
				log.Println(u)
			}

			r, err := p.getter.Get(u)
			if err != nil {
				return nil, err
			}
			/*
				b, _ := ioutil.ReadAll(r)
				f, _ := os.Create("coll.json")

				f.Write(b)
				f.Close()
			*/
			var result APIResult
			err = json.NewDecoder(r).Decode(&result)
			if err != nil {
				return nil, err
			}

			if len(result.Data) == 0 {
				// A collection of collection (a serie, indeed) enrty hasn't any Data. We have to fetch collections for each season
				if seasonSearched {
					return nil, errors.New("Collection not found " + d.ProgramID)
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
					return nil, err
				}
				continue collectionLoop
			}

			for _, ep := range result.Data {
				show := &providers.Show{
					ID:      ep.ProgramID,
					Show:    d.Title, //Takes collection's title
					Title:   ep.Subtitle,
					Pitch:   ep.ShortDescription,
					ShowURL: ep.URL,
					Season:  seasons[0],
				}

				img := getBestImage(ep.Images, "square")
				if len(img) == 0 {
					img = getBestImage(ep.Images, "landscape")
				}
				show.ThumbnailURL = img
				err := p.GetShowInfo(show)
				if err != nil {
					log.Println(err)
					continue
				}
				setEpisodeFormTitle(show, ep.Title)
				shows = append(shows, show)
			}

			u = result.NextPage

		}
		delete(collectionIDs, seasons[0]) // Season on top of the stack is done.
	}

	return shows, nil
}

var (
	parseTitleSeasonEpisode = regexp.MustCompile(`([^-]+) - Saison (\d+) \((\d+)\/\d+\)`)
	parseTitle              = regexp.MustCompile(`([^-]+) - (.+)`)
)

func setEpisodeFormTitle(show *providers.Show, t string) {

	m := parseTitleSeasonEpisode.FindAllStringSubmatch(t, -1)
	if len(m) > 0 {
		show.Show = m[0][1]
		show.Season = m[0][2]
		show.Episode = m[0][3]
		return
	}
	m = parseTitleSeasonEpisode.FindAllStringSubmatch(show.Title, -1)
	if len(m) > 0 {
		show.Show = m[0][1]
		show.Season = ""
		show.Episode = ""
		return
	}
	if show.Title == "" {
		show.Title = t
	}
}

// getBestImage retreive the url for the image of type "protrait/banner/landscape..." with the highest resolution
func getBestImage(images Images, t string) string {
	image, ok := images[t]
	if !ok {
		return ""
	}

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

// GetShowFileName return a file name with a path that is compatible with PLEX server:
//   ShowName/Season NN/ShowName - sNNeMM - Episode title
//   Show and Episode names are sanitized to avoid problem when saving on the file system
func (p *ArteTV) GetShowFileName(s *providers.Show) string {
	err := p.GetShowInfo(s)
	if err != nil {
		return ""
	}
	var showPath, seasonPath, episodePath string
	showPath = providers.PathNameCleaner(s.Show)

	if s.Season == "" {
		seasonPath = "Season " + s.AirDate.Format("2006")
	} else {
		seasonPath = "Season " + providers.Format2Digits(s.Season)
	}

	if s.Episode == "" {
		episodePath = providers.FileNameCleaner(s.Show) + " - " + s.AirDate.Format("2006-01-02")
	} else {
		episodePath = providers.FileNameCleaner(s.Show) + " - s" + providers.Format2Digits(s.Season) + "e" + providers.Format2Digits(s.Episode)
	}

	if s.Title == "" || s.Title == s.Show {
		episodePath += s.ID + ".mp4"
	} else {
		episodePath += " - " + providers.FileNameCleaner(s.Title) + ".mp4"
	}

	return filepath.Join(showPath, seasonPath, episodePath)
}

// GetShowFileNameMatcher return a file pattern of this show
// used for detecting already got episode even when episode or season is different
func (p *ArteTV) GetShowFileNameMatcher(s *providers.Show) string {
	return p.GetShowFileName(s)
}

// https://api.arte.tv/api/player/v1/config/fr/083668-012-A?autostart=1&lifeCycle=1

const arteDetails = "https://api.arte.tv/api/player/v1/config/fr/%s?autostart=1&lifeCycle=1" // Player to get Video streams ProgID

// GetShowStreamURL return the show's URL, a mp4 file
func (p *ArteTV) GetShowStreamURL(s *providers.Show) (string, error) {
	if s.StreamURL != "" {
		return s.StreamURL, nil
	}

	err := p.GetShowInfo(s)
	if err != nil {
		return "", err
	}

	return s.StreamURL, nil
}

// GetShowInfo gather show information from dedicated web page.
// It load the html page of the show to extract availability date used as airdate and production year as season
func (p *ArteTV) GetShowInfo(s *providers.Show) error {
	if s.Detailed {
		return nil
	}

	url := fmt.Sprintf(arteDetails, s.ID)
	if p.debug {
		log.Println(url)
	}
	r, err := p.getter.Get(url)
	if err != nil {
		return fmt.Errorf("Can't get show's detailled information: %v", err)
	}
	player := playerAPI{}
	err = json.NewDecoder(r).Decode(&player)
	if err != nil {
		return fmt.Errorf("Can't decode show's detailled information: %v", err)
	}

	s.StreamURL = p.getBestVideo(player.VideoJSONPlayer.VSR)
	s.AirDate = time.Time(player.VideoJSONPlayer.VRA)
	s.Detailed = true

	return nil
}

type mapStrInt map[string]uint64

// getBestVideo return the best video stream given preferences
//   Streams are scored in following order:
//   - Version (VF,VF_ST) that match preference
//   - Stream quality, the highest possible
//   - Preferred format
// The URL's stream with the best score is returned
func (p *ArteTV) getBestVideo(ss map[string]StreamInfo) string {
	scores := mapStrInt{}
	sortedResolution := getPlayerResolutions(ss)

	for k, s := range ss {
		scores[k] = p.getStreamScore(s, reverseSliceIndex(getResolutionKey(s), sortedResolution))
	}

	scoreSlice := sortMapStrInt(scores)
	return ss[scoreSlice[0]].URL
}

func getPlayerResolutions(ss map[string]StreamInfo) []string {
	scoreResolution := mapStrInt{}
	for _, s := range ss {
		p := uint64(s.Height) * uint64(s.Width) * uint64(s.Bitrate)
		scoreResolution[getResolutionKey(s)] = p
	}
	return sortMapStrInt(scoreResolution)
}

func getResolutionKey(s StreamInfo) string {
	return strconv.Itoa(s.Width) + "*" + strconv.Itoa(s.Height) + "*" + strconv.Itoa(s.Bitrate)
}

func (p *ArteTV) getStreamScore(s StreamInfo, resolutionIndex uint64) uint64 {
	grade := uint64(0)

	// Best grade for the preferred version
	grade += reverseSliceIndex(s.VersionCode, p.preferredVersions) * 1000000

	// Then best resolution
	grade += resolutionIndex * 1000

	// Add points for the preferred format
	if s.MediaType == p.preferredMedia {
		grade += 10
	}
	return grade
}

// sortMapStrInt return a slice of string in the order int
func sortMapStrInt(m mapStrInt) []string {
	type kv struct {
		k string
		v uint64
	}
	s := make([]kv, len(m))
	i := 0
	for k, v := range m {
		s[i] = kv{k: k, v: v}
		i++
	}
	sort.Slice(s, func(i, j int) bool {
		return s[i].v > s[j].v
	})
	r := make([]string, len(m))
	for i, v := range s {
		r[i] = v.k
	}
	return r
}

func sliceIndex(k string, ls []string) uint64 {
	for i, s := range ls {
		if s == k {
			return uint64(i + 1)
		}
	}
	return 0
}

func reverseSliceIndex(k string, ls []string) uint64 {
	r := sliceIndex(k, ls)
	if r == 0 {
		return r
	}
	return uint64(len(ls)+1) - r

}

func episodeFromID(ID string) string {
	b := ""
	for _, c := range ID {
		if c >= '0' && c <= '9' {
			b += string(c)
		}
	}
	for len(b) > 0 && b[0] == '0' {
		b = b[1:]
	}
	return b
}
