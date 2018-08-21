package artetv

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/simulot/aspiratv/parsers/jscript"

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
	arteGuide      = "https://www.arte.tv/guide/api/api/pages/fr/TV_GUIDE/?day="                           // Day YY-MM-DD
	arteDetails    = "https://api.arte.tv/api/player/v1/config/fr/%s?autostart=1&lifeCycle=1"              // Player to get Video streams ProgID
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
}

// WithGetter inject a getter in FranceTV object instead of normal one
func WithGetter(g getter) func(p *ArteTV) {
	return func(p *ArteTV) {
		p.getter = g

	}
}

// New setup a Show provider for Arte
func New(conf ...func(p *ArteTV)) (*ArteTV, error) {
	p := &ArteTV{
		getter: http.DefaultClient,
		//TODO: get preferences from config file
		preferredVersions: []string{"VF", "VOF", "VOF-STF", "VOSTF", "VF-STF"}, // "VF-STMF" "VA", "VA-STA"
		preferredMedia:    "mp4",
		htmlParserFactory: htmlparser.NewFactory(),
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
	log.Print("[artetv] Fetch Arte's new shows")

	replay, err := p.getReplayShows(mm)
	if err != nil {
		return nil, err
	}
	shows = append(shows, replay...)

	collections, err := p.getCollectionsShows(mm)
	if err != nil {
		return nil, err
	}

	shows = append(shows, collections...)

	return shows, nil
}

// browse all match request to handle those having a ShowID denoting an Arte collection ID
// and invoke  getCollection
func (p *ArteTV) getCollectionsShows(mm []*providers.MatchRequest) ([]*providers.Show, error) {
	shows := []*providers.Show{}
	for _, m := range mm {
		if m.Provider == "artetv" && m.Playlist != "" {
			collection, err := p.getCollection(m.Playlist, m.Destination)
			if err != nil {
				log.Printf("[artetv] Can't fetch collection %q: %v", m.Playlist, err)
				continue
			}
			shows = append(shows, collection...)
		}
	}
	return shows, nil
}

// getCollectionIDFromName retrieve collection's ID from its name
// It returns the 1st encountered collection in result set
func (p *ArteTV) getCollectionIDFromName(collection string) (string, error) {

	if p.debug {
		log.Printf("[artetv] Query collection's ID: %q", collection)
	}

	URL := fmt.Sprintf(arteSearch, url.PathEscape(collection))
	r, err := p.getter.Get(URL)
	if err != nil {
		return "", err
	}
	d := json.NewDecoder(r)
	result := &searchResults{}
	err = d.Decode(result)
	if err != nil {
		return "", err
	}
	for _, s := range result.Data {
		if strings.HasPrefix(s.ProgramID, "RC-") {
			return s.ProgramID, nil
		}
	}
	return "", fmt.Errorf("[artetv] Id for collection %q not found", collection)
}

// get all Arte shows for the given collection ID
func (p *ArteTV) getCollection(ColName string, destination string) ([]*providers.Show, error) {

	ColID, err := p.getCollectionIDFromName(ColName)
	if err != nil {
		return nil, err
	}

	shows := []*providers.Show{}

	if p.debug {
		log.Printf("[artetv] Fetch collection: %q[%s]", ColName, ColID)
	}
	page := 1

	for {
		URL := fmt.Sprintf(arteCollection, ColID, page)
		r, err := p.getter.Get(URL)
		if err != nil {
			return nil, err
		}
		d := json.NewDecoder(r)
		collection := &searchResults{}
		err = d.Decode(collection)
		if err != nil {
			return nil, err
		}
		for _, data := range collection.Data {
			s := &providers.Show{
				AirDate:  time.Time{},
				Channel:  "Arte",
				Category: "",
				Detailed: false,
				DRM:      false,
				Duration: data.Duration.Duration(),
				Episode:  "",
				ID:       data.ProgramID,
				Pitch:    strings.TrimSpace(data.Description),
				Season:   "",
				Provider: "artetv",
				Show: func() string {
					if len(data.Subtitle) == 0 {
						return strings.TrimSpace(collection.Link.Title)
					}
					return strings.TrimSpace(data.Title)
				}(),
				ShowURL:   data.URL,
				StreamURL: "", // Must call GetShowStreamURL to get the show's URL
				ThumbnailURL: func(t thumbs) string {
					bestRes := -1
					bestURL := ""
					for _, r := range t.Resolutions {
						if r.Height*r.Width > bestRes {
							bestRes = r.Height * r.Width
							bestURL = r.URL
						}
					}
					return bestURL
				}(data.Images["landscape"]),
				Title: func() string {
					if len(data.Subtitle) == 0 {
						return strings.TrimSpace(data.Title)
					}
					return strings.TrimSpace(data.Subtitle)
				}(),
				Destination: destination,
			}
			setEpisodeFromTitle(s)
			shows = append(shows, s)
		}
		if len(collection.NextPage) == 0 {
			break
		}
		page++
	}
	if p.debug {
		log.Printf("[artetv] Collection: %q has %d shows", ColName, len(shows))
	}
	return shows, nil
}

// parse guide page to get brocasted shows.
// Note for the first run (in service mode), we collect 3 weeks of guide
func (p *ArteTV) getReplayShows(mm []*providers.MatchRequest) ([]*providers.Show, error) {
	var dateStart time.Time

	shows := []*providers.Show{}
	if runCounter == 0 {
		// Start search 3 weeks in the past
		dateStart = time.Now().Truncate(24 * time.Hour).Add(-3 * 7 * 24 * time.Hour)
	} else {
		// Start today
		dateStart = time.Now().Truncate(24 * time.Hour)
	}

	dateEnd := time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour)

	for d := dateStart; d.Before(dateEnd); d = d.Add(24 * time.Hour) {
		ss, err := p.getGuide(mm, d)
		if err != nil {
			return nil, err
		}
		shows = append(shows, ss...)
	}
	runCounter++
	return shows, nil
}

// getGuide get Arte's guide of programs for the given date
func (p *ArteTV) getGuide(mm []*providers.MatchRequest, d time.Time) ([]*providers.Show, error) {
	if p.debug {
		log.Printf("[artetv] Fetch guide for date: %s", d.Format("06-01-02"))
	}
	url := arteGuide + d.Format("06-01-02")
	r, err := p.getter.Get(url)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(r)
	guide := &guide{}
	err = decoder.Decode(guide)
	if err != nil {
		return nil, err
	}

	shows := []*providers.Show{}

	for _, z := range guide.Zones {
		if z.Code.Name == "listing_TV_GUIDE" {
			for _, d := range z.Data {
				s := &providers.Show{
					AirDate: func(ds []tsGuide) time.Time {
						if len(ds) > 0 {
							return ds[0].Time()
						}
						return time.Time{}
					}(d.BroadcastDates),
					Channel:   "Arte",
					Category:  "",
					Detailed:  false,
					DRM:       false,
					Duration:  d.Duration.Duration(),
					Episode:   "",
					ID:        d.ProgramID,
					Pitch:     strings.TrimSpace(d.ShortDescription),
					Season:    "",
					Show:      strings.TrimSpace(d.Title),
					Provider:  "artetv",
					ShowURL:   d.URL,
					StreamURL: "", // Must call GetShowStreamURL to get the show's URL
					ThumbnailURL: func(t thumbs) string {
						bestRes := -1
						bestURL := ""
						for _, r := range t.Resolutions {
							if r.Height*r.Width > bestRes {
								bestRes = r.Height * r.Width
								bestURL = r.URL
							}
						}
						return bestURL
					}(d.Images["landscape"]),
					Title: strings.TrimSpace(d.Subtitle),
				}
				setEpisodeFromTitle(s)
				if providers.IsShowMatch(mm, s) {
					shows = append(shows, s)
				}
			}
		}
	}
	return shows, nil
}

var reArteSeries = regexp.MustCompile(`(?P<Title>.*\S)\s*\((?P<Episode>\d+)\/(?P<Total>\d+)\)`)

// Get episode number from the title pattern (episode/number of episodes) in the title
// If found, the pattern (x/y) is removed from title
func setEpisodeFromTitle(s *providers.Show) {
	m := reArteSeries.FindAllStringSubmatch(s.Title, -1)
	if m != nil {
		s.Title = m[0][1]
		s.Episode = m[0][2]
	}
}

// GetShowStreamURL return the show's URL, a mp4 file
func (p *ArteTV) GetShowStreamURL(s *providers.Show) (string, error) {
	if s.StreamURL != "" {
		return s.StreamURL, nil
	}

	if p.debug {
		log.Printf("Fetch video url for %q, %q", s.Show, s.Title)
	}
	url := fmt.Sprintf(arteDetails, s.ID)
	r, err := p.getter.Get(url)
	if err != nil {
		return "", fmt.Errorf("Can't get show's detailled information: %v", err)
	}

	d := json.NewDecoder(r)
	i := &player{}
	err = d.Decode(&i)
	if err != nil {
		return "", fmt.Errorf("Can't decode show's detailled information: %v", err)
	}

	s.StreamURL = p.getBestVideo(i.VideoJSONPlayer.VSR)

	return s.StreamURL, nil
}

// GetShowInfo gather show information from dedicated web page.
// It load the html page of the show to extract availability date used as airdate and production year as season
func (p *ArteTV) GetShowInfo(s *providers.Show) error {
	if s.Detailed {
		return nil
	}
	r, err := p.getter.Get(s.ShowURL)
	if err != nil {
		return err
	}

	info, err := readDetails(r)
	if err != nil {
		return err
	}
	s.AirDate = info.airDate
	s.Season = info.season
	return nil
}

type showInfo struct {
	season   string
	airDate  time.Time
	title    string
	subTitle string
}

// readDetails returns the structure that contains shows details

func readDetails(r io.Reader) (*showInfo, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	info := &showInfo{}

	o, err := jscript.ParseObjectAtAnchor(b, regexp.MustCompile(`"zones":\[\{`))
	if err != nil {
		return nil, err
	}

	if dd := o.Property("data"); dd != nil {
		for _, d := range dd.Ar {
			if t := d.Property("title"); t != nil {
				info.title = strings.TrimSpace(t.String())
			}
			if t := d.Property("subtitle"); t != nil {
				info.subTitle = strings.TrimSpace(t.String())
			}
			if a := d.Property("availability"); a != nil {
				if t := a.Property("startDay"); t != nil {
					d, err := time.Parse("2006-01-02", t.String())
					if err == nil {
						info.airDate = d
					}
				}
			}
			if cc := d.Property("credits"); cc != nil {
				for _, c := range cc.Ar {
					if code := c.Property("code"); code != nil && code.String() == "PRODUCTION_YEAR" {
						y := c.Property("values").Strings()
						info.season = y[0]
					}
				}
			}
		}
	}

	return info, nil
}

type mapStrInt map[string]uint64

// getBestVideo return the best video stream given preferences
//   Streams are scored in following order:
//   - Version (VF,VF_ST) that match preference
//   - Stream quality, the highest possible
//   - Preferred format
// The URL's stream with the best score is returned
func (p *ArteTV) getBestVideo(ss map[string]streamInfo) string {
	scores := mapStrInt{}
	sortedResolution := getPlayerResolutions(ss)

	for k, s := range ss {
		scores[k] = p.getStreamScore(s, reverseSliceIndex(getResolutionKey(s), sortedResolution))
	}

	scoreSlice := sortMapStrInt(scores)
	return ss[scoreSlice[0]].URL
}

func getPlayerResolutions(ss map[string]streamInfo) []string {
	scoreResolution := mapStrInt{}
	for _, s := range ss {
		p := uint64(s.Height) * uint64(s.Width) * uint64(s.Bitrate)
		scoreResolution[getResolutionKey(s)] = p
	}
	return sortMapStrInt(scoreResolution)
}

func getResolutionKey(s streamInfo) string {
	return strconv.Itoa(s.Width) + "*" + strconv.Itoa(s.Height) + "*" + strconv.Itoa(s.Bitrate)
}

func (p *ArteTV) getStreamScore(s streamInfo, resolutionIndex uint64) uint64 {
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

// GetShowFileName return a file name with a path that is compatible with PLEX server:
//   ShowName/Season NN/ShowName - sNNeMM - Episode title
//   Show and Episode names are sanitized to avoid problem when saving on the file system
func (p *ArteTV) GetShowFileName(s *providers.Show) string {
	if !s.Detailed {
		p.GetShowInfo(s)
	}
	switch {
	case s.Season == "" && s.Episode == "" && s.AirDate.IsZero():
		// Following Plex naming convention for Specials show https://support.plex.tv/articles/200220707-naming-tv-show-specials/
		return filepath.Join(
			providers.PathNameCleaner(s.Show),
			"Specials",
			providers.FileNameCleaner(s.Show)+" - s00e"+episodeFromID(s.ID)+" - "+providers.FileNameCleaner(s.Title)+".mp4",
		)
	case s.Season == "" && s.Episode == "" && !s.AirDate.IsZero():
		// Follow Plex naming convention https://support.plex.tv/articles/200381053-naming-date-based-tv-shows/
		return filepath.Join(
			providers.PathNameCleaner(s.Show),
			"Season "+strconv.Itoa(s.AirDate.Year()),
			providers.FileNameCleaner(s.Show)+" - s"+strconv.Itoa(s.AirDate.Year())+"e"+episodeFromID(s.ID)+" - "+providers.FileNameCleaner(s.Title)+".mp4",
		)
	case s.Season != "" && s.Episode == "" && !s.AirDate.IsZero():
		// When episode is missing, use the ID as episode number
		return filepath.Join(
			providers.PathNameCleaner(s.Show),
			"Season "+providers.Format2Digits(s.Season),
			providers.FileNameCleaner(s.Show)+" - s"+s.Season+"e"+episodeFromID(s.ID)+" - "+providers.FileNameCleaner(s.Title)+".mp4",
		)
	}
	// Normal case: https://support.plex.tv/articles/200220687-naming-series-season-based-tv-shows/
	return filepath.Join(
		providers.PathNameCleaner(s.Show),
		"Season "+providers.Format2Digits(s.Season),
		providers.FileNameCleaner(s.Show)+" - s"+providers.Format2Digits(s.Season)+"e"+providers.Format2Digits(s.Episode)+" - "+providers.FileNameCleaner(s.Title)+".mp4",
	)
}

// GetShowFileNameMatcher return a file pattern of this show
// used for detecting already got episode even when episode or season is different
func (ArteTV) GetShowFileNameMatcher(s *providers.Show) string {
	return filepath.Join(
		providers.PathNameCleaner(s.Show),
		"*",
		providers.FileNameCleaner(s.Show)+" - * - "+providers.FileNameCleaner(s.Title)+".mp4",
	)
}
