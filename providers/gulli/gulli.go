package gulli

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/simulot/aspiratv/net/myhttp"
	"github.com/simulot/aspiratv/parsers/htmlparser"
	"github.com/simulot/aspiratv/providers"
)

type getter interface {
	Get(ctx context.Context, uri string) (io.ReadCloser, error)
}

// Gulli provider gives access to Gulli catchup tv
type Gulli struct {
	getter            getter
	htmlParserFactory *htmlparser.Factory
	seenShows         map[string]bool
	debug             bool
	cacheFile         string
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
func New(conf ...func(p *Gulli)) (*Gulli, error) {

	p := &Gulli{
		getter:            myhttp.DefaultClient,
		htmlParserFactory: nil,
		seenShows:         map[string]bool{},
	}
	for _, f := range conf {
		f(p)
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

// DebugMode set debug mode
func (p *Gulli) DebugMode(b bool) {
	p.debug = b
}

// withGetter set a getter for Gulli
func withGetter(g getter) func(p *Gulli) {
	return func(p *Gulli) {
		p.getter = g
	}
}

// Name return the name of the provider
func (p Gulli) Name() string { return "gulli" }

// Shows download the shows catalog from the web site.
func (p *Gulli) Shows(ctx context.Context, mm []*providers.MatchRequest) chan *providers.Show {
	shows := make(chan *providers.Show)

	go func() {
		defer close(shows)
		cat, err := p.downloadCatalog(ctx)
		if err != nil {
			log.Printf("[%s] Can't call replay catalog: %q", p.Name(), err)
			return
		}

		for _, s := range cat {
			for _, m := range mm {
				if strings.Contains(strings.ToLower(s.Title), m.Show) {
					ID, err := p.getFirstEpisodeID(ctx, s)
					showTitles, err := p.getPlayer(ctx, ID)
					if err != nil {
						log.Printf("[%s] Can't decode replay catalog: %q", p.Name(), err)
						return
					}
					for _, t := range showTitles {
						t.Destination = m.Destination
						shows <- t
					}
				}
			}
		}
	}()
	return shows
}

// GetShowStreamURL return the show's URL, a mp4 file
func (p *Gulli) GetShowStreamURL(ctx context.Context, s *providers.Show) (string, error) {
	return s.StreamURL, nil
}

// GetShowInfo gather show information from dedicated web page.
// It load the html page of the show to extract availability date used as airdate and production year as season
func (p *Gulli) GetShowInfo(ctx context.Context, s *providers.Show) error {
	return nil
}
