package francetv

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/simulot/aspiratv/net/myhttp/httptest"
	"github.com/simulot/aspiratv/net/myhttp"
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
	WSInfoOeuvre = "http://webservices.francetelevisions.fr/tools/getInfosOeuvre/v2/?catalogue=Pluzz&idDiffusion=" // Show's video link and details
)

type getter interface {
	Get(ctx context.Context, uri string) (io.ReadCloser, error)
	DoWithContext(ctx context.Context, method string, theURL string, headers http.Header, body io.Reader) (io.ReadCloser, error)
}

// FranceTV structure handles france-tv catalog of shows
type FranceTV struct {
	getter   getter
	debug    bool
	deadline time.Duration
	algolia  *AlgoliaConfig
}

// WithGetter inject a getter in FranceTV object instead of normal one
func WithGetter(g getter) func(ftv *FranceTV) {
	return func(ftv *FranceTV) {
		ftv.getter = g
	}
}

// New setup a Show provider for France Télévisions
func New(conf ...func(ftv *FranceTV)) (*FranceTV, error) {
	p := &FranceTV{
		getter: myhttp.DefaultClient,
	}
	p.deadline = 30 * time.Second

	for _, fn := range conf {
		fn(p)
	}
	return p, nil
}

// Name return the name of the provider
func (FranceTV) Name() string { return "francetv" }

// DebugMode switch debug mode
func (p *FranceTV) DebugMode(mode bool) {
	p.debug = mode
	if mode {
		p.deadline = time.Hour
	} else {
		p.deadline = 30 * time.Second
	}
}

// Shows return shows that match with matching list.
func (p *FranceTV) Shows(ctx context.Context, mm []*providers.MatchRequest) chan *providers.Show {
	err := p.getAlgoliaConfig(ctx)

	if err != nil {
		return nil
	}
	shows := make(chan *providers.Show)

	go func() {
		defer close(shows)
		for _, m := range mm {
			if m.Provider != "francetv" {
				continue
			}
			for s := range p.queryAlgolia(ctx, m) {
				shows <- s
			}
		}
	}()
	return shows
}

// GetShowStreamURL return the show's URL, a m3u8 playlist
func (p *FranceTV) GetShowStreamURL(ctx context.Context, s *providers.Show) (string, error) {
	if s.StreamURL == "" {
		err := p.GetShowInfo(ctx, s)
		if err != nil {
			return "", fmt.Errorf("Can't get detailed information for the show: %v", err)
		}
	}
	return s.StreamURL, nil
}

// GetShowInfo query the URL from InfoOeuvre web service
func (p *FranceTV) GetShowInfo(ctx context.Context, s *providers.Show) error {
	if s.Detailed {
		return nil
	}
	i := infoOeuvre{}

	url := WSInfoOeuvre + s.ID
	if p.debug {
		log.Printf("[%s] Get details url: %q", p.Name(), url)
	}
	r, err := p.getter.Get(ctx, url)
	if err != nil {
		return fmt.Errorf("Can't get show's detailed information: %v", err)
	}
	if p.debug {
		r = httptest.DumpReaderToFile(r, "francetv-info-"+s.ID+"-")
	}

	err = json.NewDecoder(r).Decode(&i)
	if err != nil {
		return fmt.Errorf("Can't decode show's detailed information: %v", err)
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}

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
func (FranceTV) GetShowFileName(ctx context.Context, s *providers.Show) string {

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

	if s.Episode == "" && (s.Title == "" || s.Title == s.Show) {
		episodePath += " - " + s.ID + ".mp4"
	} else {
		if s.Title != "" && s.Title != s.Show {
			episodePath += " - " + providers.FileNameCleaner(s.Title) + ".mp4"
		} else {
			episodePath += ".mp4"
		}
	}

	return filepath.Join(showPath, seasonPath, episodePath)

}

// GetShowFileNameMatcher return a file pattern of this show
// used for detecting already got episode even when episode or season is different
func (p *FranceTV) GetShowFileNameMatcher(ctx context.Context, s *providers.Show) string {
	return p.GetShowFileName(ctx, s)
}
