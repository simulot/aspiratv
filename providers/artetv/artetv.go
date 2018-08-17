package artetv

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/simulot/aspiratv/net/http"
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
	arteIndex   = "https://www.arte.tv"
	arteCDN     = "https://static-cdn.arte.tv"
	arteGuide   = "https://www.arte.tv/guide/api/api/pages/fr/TV_GUIDE/?day="
	arteDetails = "https://api.arte.tv/api/player/v1/config/fr/%s?autostart=1&lifeCycle=1"
)

// Track when this is the first time the Show is invoked
var runCounter = 0

type getter interface {
	Get(uri string) (io.Reader, error)
}

// ArteTV structure handles arte  catalog of shows
type ArteTV struct {
	getter           getter
	preferedVersions []string // versionCode List of version in order of preference VF,VA...
	preferedMedia    string   // mediaType mp4,hls
	debug            bool
}

// WithGetter inject a getter in FranceTV object instead of normal one
func WithGetter(g getter) func(p *ArteTV) {
	return func(p *ArteTV) {
		p.getter = g

	}
}

// New setup a Show provider for France Télévisons
func New(conf ...func(p *ArteTV)) (*ArteTV, error) {
	p := &ArteTV{
		getter: http.DefaultClient,
		//TODO: get prefrences from config file
		preferedVersions: []string{"VF", "VOF", "VOF-STF", "VOSTF", "VF-STF"}, // "VF-STMF" "VA", "VA-STA"
		preferedMedia:    "mp4",
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
func (p *ArteTV) Shows() ([]*providers.Show, error) {
	shows := []*providers.Show{}

	var dateStart time.Time

	if runCounter == 0 {
		// Start search 3 weeks in the past
		dateStart = time.Now().Truncate(24 * time.Hour).Add(-3 * 7 * 24 * time.Hour)
	} else {
		// Start today
		dateStart = time.Now().Truncate(24 * time.Hour)
	}

	dateEnd := time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour)

	for d := dateStart; d.Before(dateEnd); d = d.Add(24 * time.Hour) {
		ss, err := p.getGuide(d)
		if err != nil {
			return nil, err
		}
		shows = append(shows, ss...)
	}
	return shows, nil
}

// getGuide get Arte's guide of programs for the given date
func (p *ArteTV) getGuide(d time.Time) ([]*providers.Show, error) {
	if p.debug {
		log.Printf("Fetch guide for date: %s", d.Format("06-01-02"))
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
				shows = append(shows, &providers.Show{
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
				})
			}
		}
	}
	return shows, nil
}

// GetShowStreamURL return the show's URL, a m3u8 playlist
func (p *ArteTV) GetShowStreamURL(s *providers.Show) (string, error) {
	if s.StreamURL == "" {
		err := p.GetShowInfo(s)
		if err != nil {
			return "", err
		}
	}
	return s.StreamURL, nil
}

var reArteSerie = regexp.MustCompile(`(?P<Title>.*\S)\s*\((?P<Episode>\d+)\/(?P<Total>\d+)\)`)

// GetShowInfo query the URL from InfoOeuvre web service
func (p *ArteTV) GetShowInfo(s *providers.Show) error {
	if s.Detailed {
		return nil
	}
	if p.debug {
		log.Printf("Fetch details for %q, %q", s.Show, s.Title)
	}
	url := fmt.Sprintf(arteDetails, s.ID)
	r, err := p.getter.Get(url)
	if err != nil {
		return fmt.Errorf("Can't get show's detailled information: %v", err)
	}

	d := json.NewDecoder(r)
	i := &player{}
	err = d.Decode(&i)
	if err != nil {
		return fmt.Errorf("Can't decode show's detailled information: %v", err)
	}

	// Get episode number from the title when exists.
	m := reArteSerie.FindAllStringSubmatch(i.VideoJSONPlayer.VTI, -1)
	if m != nil {
		s.Title = m[0][1]
		s.Episode = m[0][2]
		s.Season = strconv.Itoa(s.AirDate.Year())
	} else {
		s.Title = i.VideoJSONPlayer.VTI
		s.Episode = s.AirDate.Format("2006-01-02")
		s.Season = strconv.Itoa(s.AirDate.Year())
	}

	s.StreamURL = p.getBestVideo(i.VideoJSONPlayer.VSR)

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
	grade += reverseSliceIndex(s.VersionCode, p.preferedVersions) * 1000000

	// Then best resolution
	grade += resolutionIndex * 1000

	// Add points for the preferred format
	if s.MediaType == p.preferedMedia {
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

// GetShowFileName return a file name with a path that is compatible with PLEX server:
//   ShowName/Season NN/ShowName - sNNeMM - Episode title
//   Show and Episode names are sanitized to avoid problem when saving on the file system
func (ArteTV) GetShowFileName(s *providers.Show) string {
	if s.Season == "" && s.Episode == "" {
		// Follow Plex naming convention https://support.plex.tv/articles/200381053-naming-date-based-tv-shows/
		return filepath.Join(
			providers.PathNameCleaner(s.Show),
			"Season "+strconv.Itoa(s.AirDate.Year()),
			providers.FileNameCleaner(s.Show)+" - "+s.AirDate.Format("2006-01-02")+" - "+providers.FileNameCleaner(s.Title)+".mp4",
		)
	}
	if s.Season != "" && s.Episode == "" {
		// When episode is missing, use the ID as episode number
		return filepath.Join(
			providers.PathNameCleaner(s.Show),
			"Season "+providers.Format2Digits(s.Season),
			providers.FileNameCleaner(s.Show)+" - s"+providers.Format2Digits(s.Season)+"e"+s.ID+" - "+providers.FileNameCleaner(s.Title)+".mp4",
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
		"Season *",
		providers.FileNameCleaner(s.Show)+" * "+providers.FileNameCleaner(s.Title)+".mp4",
	)
}
