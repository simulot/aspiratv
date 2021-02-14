package gulli

import (
	"context"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/simulot/aspiratv/matcher"
	"github.com/simulot/aspiratv/media"
	"github.com/simulot/aspiratv/metadata/nfo"
	"github.com/simulot/aspiratv/net/myhttp"
	"github.com/simulot/aspiratv/parsers/htmlparser"
	"github.com/simulot/aspiratv/providers"
)

type getter interface {
	Get(ctx context.Context, uri string) (io.ReadCloser, error)
}

// Gulli provider gives access to Gulli catchup tv
type Gulli struct {
	config            providers.Config
	getter            getter
	htmlParserFactory *htmlparser.Factory
	seenShows         map[string]bool
	cacheFile         string
	deadline          time.Duration
	cartoonList       []ShowEntry
	tvshows           map[string]*nfo.TVShow
}

// init registers Gulli provider
func init() {
	p, err := New()
	if err != nil {
		panic(err)
	}
	providers.Register(p)
}

// New creates a Gulli provider with given configuration
func New() (*Gulli, error) {

	p := &Gulli{
		getter:            myhttp.DefaultClient,
		htmlParserFactory: nil,
		seenShows:         map[string]bool{},
		deadline:          30 * time.Second,
		tvshows:           map[string]*nfo.TVShow{},
	}
	if rt, ok := p.getter.(http.RoundTripper); ok {
		p.htmlParserFactory = htmlparser.NewFactory(htmlparser.SetTransport(rt))
	} else {
		p.htmlParserFactory = htmlparser.NewFactory()
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	p.cacheFile = path.Join(cacheDir, "aspiratv", "gulli-catalog.json")
	return p, nil
}

func (p *Gulli) Configure(c providers.Config) {
	p.config = c
	if p.config.Log.IsDebug() {
		p.deadline = time.Hour
	}
}

// withGetter set a getter for Gulli
func withGetter(g getter) func(p *Gulli) {
	return func(p *Gulli) {
		p.getter = g
	}
}

// Name return the name of the provider
func (p Gulli) Name() string { return "gulli" }

// MediaList download the shows catalog from the web site.
func (p *Gulli) MediaList(ctx context.Context, mm []*matcher.MatchRequest) chan *media.Media {
	shows := make(chan *media.Media)

	go func() {
		defer close(shows)
		cat, err := p.downloadCatalog(ctx)
		if err != nil {
			p.config.Log.Error().Printf("[%s] Can't call replay catalog: %q", p.Name(), err)
			return
		}

		for _, s := range cat {
			for _, m := range mm {
				if strings.Contains(strings.ToLower(s.Title), m.Show) {
					ID, err := p.getFirstEpisodeID(ctx, s)
					showTitles, err := p.getPlayer(ctx, m, ID)
					if err != nil {
						p.config.Log.Error().Printf("[%s] Can't decode replay catalog: %q", p.Name(), err)
						return
					}
					for _, s := range showTitles {
						shows <- s
					}
				}
			}
		}
	}()
	return shows
}

// GetMediaDetails gather show information from dedicated web page.
func (p *Gulli) GetMediaDetails(ctx context.Context, m *media.Media) error {

	return nil
}
