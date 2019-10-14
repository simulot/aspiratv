package gulli

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

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
	getter            getter
	htmlParserFactory *htmlparser.Factory
	seenShows         map[string]bool
	debug             bool
	cacheFile         string
	deadline          time.Duration
	cartoonList       []ShowEntry
	tvshows           map[string]*nfo.TVShow
	keepBonuses       bool
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
	p.keepBonuses = c.KeepBonus
	p.debug = c.Debug
	if p.debug {
		p.deadline = time.Hour
	} else {
		p.deadline = 30 * time.Second
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
func (p *Gulli) MediaList(ctx context.Context, mm []*providers.MatchRequest) chan *providers.Media {
	shows := make(chan *providers.Media)

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
					showTitles, err := p.getPlayer(ctx, m, ID)
					if err != nil {
						log.Printf("[%s] Can't decode replay catalog: %q", p.Name(), err)
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
func (p *Gulli) GetMediaDetails(ctx context.Context, m *providers.Media) error {
	var err error
	info := m.Metadata.GetMediaInfo()
	title := strings.ToLower(info.Showtitle)
	if info.TVShow == nil {
		tvshow, ok := p.tvshows[title]
		if !ok {
			tvshow, err = p.getShowInfo(ctx, title)
			if err != nil {
				return err
			}
		}

		info.TVShow = tvshow
		return nil

	}
	return nil
}
