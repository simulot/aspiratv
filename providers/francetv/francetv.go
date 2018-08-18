package francetv

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/simulot/aspiratv/net/http"
	"github.com/simulot/aspiratv/providers"
)

// init registers FranceTV provider
func init() {
	p, err := New()
	if err != nil {
		panic(err)
	}
	providers.Register(p)
}

// Provider constants
const (
	ProviderName = "francetv"
	WSListURL    = "http://pluzz.webservices.francetelevisions.fr/pluzz/liste/type/replay/nb/%d/debut/0"           // Available show
	WSInfoOeuvre = "http://webservices.francetelevisions.fr/tools/getInfosOeuvre/v2/?catalogue=Pluzz&idDiffusion=" // Show's video link and details
)

// Track when this is the first time the Show is invoked
var runCounter = 0

type getter interface {
	Get(uri string) (io.Reader, error)
}

// FranceTV structure handles france-tv catalog of shows
type FranceTV struct {
	getter getter
}

// WithGetter inject a getter in FranceTV object instead of normal one
func WithGetter(g getter) func(ftv *FranceTV) {
	return func(ftv *FranceTV) {
		ftv.getter = g
	}
}

// New setup a Show provider for France Télévisions
func New(conf ...func(ftv *FranceTV)) (*FranceTV, error) {
	ftv := &FranceTV{
		getter: http.DefaultClient,
	}
	for _, fn := range conf {
		fn(ftv)
	}
	return ftv, nil
}

// Name return the name of the provider
func (ftv FranceTV) Name() string { return "francetv" }

// Shows return shows that match with matching list.
func (ftv *FranceTV) Shows(mm []*providers.MatchRequest) ([]*providers.Show, error) {
	shows := []*providers.Show{}

	var url string
	if runCounter == 0 {
		url = fmt.Sprintf(WSListURL, 3000)
	} else {
		url = fmt.Sprintf(WSListURL, 500)
	}

	// Get JSON catalog of available shows on France Télévisions
	r, err := ftv.getter.Get(url)
	if err != nil {
		return nil, err
	}

	d := json.NewDecoder(r)
	list := &pluzzList{}
	err = d.Decode(list)
	if err != nil {
		return nil, fmt.Errorf("Can't decode PLUZZ list: %v", err)
	}

	for _, e := range list.Reponse.Emissions {
		// Map JSON object to provider.Show common structure
		show := &providers.Show{
			AirDate:      time.Time(e.TsDiffusionUtc),
			Channel:      e.ChaineID,
			Category:     strings.TrimSpace(e.Rubrique),
			Detailed:     false,
			DRM:          false, //TBD
			Duration:     time.Duration(e.DureeReelle),
			Episode:      e.Episode,
			ID:           e.IDDiffusion,
			Pitch:        strings.TrimSpace(e.Accroche),
			Season:       e.Saison,
			Show:         strings.TrimSpace(e.Titre),
			Provider:     ProviderName,
			ShowURL:      e.OasSitepage,
			StreamURL:    "", // Must call GetShowStreamURL to get the show's URL
			ThumbnailURL: e.ImageLarge,
			Title:        strings.TrimSpace(e.Soustitre),
		}
		if providers.IsShowMatch(mm, show) {
			shows = append(shows, show)
		}
	}
	return shows, nil
}

// GetShowStreamURL return the show's URL, a m3u8 playlist
func (ftv *FranceTV) GetShowStreamURL(s *providers.Show) (string, error) {
	if s.StreamURL == "" {
		err := ftv.GetShowInfo(s)
		if err != nil {
			return "", fmt.Errorf("Can't get detailed information for the show: %v", err)
		}
	}
	return s.StreamURL, nil
}

// GetShowInfo query the URL from InfoOeuvre web service
func (ftv *FranceTV) GetShowInfo(s *providers.Show) error {
	if s.Detailed {
		return nil
	}

	url := WSInfoOeuvre + s.ID
	r, err := ftv.getter.Get(url)
	if err != nil {
		return fmt.Errorf("Can't get show's detailled information: %v", err)
	}

	d := json.NewDecoder(r)
	i := &infoOeuvre{}
	err = d.Decode(&i)
	if err != nil {
		return fmt.Errorf("Can't decode show's detailled information: %v", err)
	}

	// May have better information than the global list
	s.Season = strconv.Itoa(i.Saison)
	s.Episode = strconv.Itoa(i.Episode)
	s.ThumbnailURL = i.ImageSecure

	for _, v := range i.Videos {
		if v.Format == "hls_v5_os" {
			s.StreamURL = v.URL
			break
		}
	}
	if s.StreamURL == "" {
		return fmt.Errorf("Can't find hls_v5_os stream for the show")
	}
	s.Detailed = true
	return nil
}

// GetShowFileName return a file name with a path that is compatible with PLEX server:
//   ShowName/Season NN/ShowName - sNNeMM - Episode title
//   Show and Episode names are sanitized to avoid problem when saving on the file system
func (FranceTV) GetShowFileName(s *providers.Show) string {
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
func (FranceTV) GetShowFileNameMatcher(s *providers.Show) string {
	return filepath.Join(
		providers.PathNameCleaner(s.Show),
		"Season *",
		providers.FileNameCleaner(s.Show)+" * "+providers.FileNameCleaner(s.Title)+".mp4",
	)
}
