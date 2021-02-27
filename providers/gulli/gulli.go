package gulli

import (
	"context"
	"strings"

	"github.com/simulot/aspiratv/matcher"
	"github.com/simulot/aspiratv/media"
	"github.com/simulot/aspiratv/metadata/nfo"
	"github.com/simulot/aspiratv/providers"
)

// Gulli provider gives access to Gulli catchup tv
type Gulli struct {
	config      providers.ProviderConfig
	seenShows   map[string]bool
	cartoonList []ShowEntry
	tvshows     map[string]*nfo.TVShow
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
		seenShows: map[string]bool{},
		tvshows:   map[string]*nfo.TVShow{},
	}
	return p, nil
}

func (p *Gulli) Configure(fns ...providers.ProviderConfigFn) {
	c := p.config
	for _, f := range fns {
		c = f(c)
	}
	p.config = c
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
			p.config.Log.Error().Printf("[%s] Can't get replay catalog: %q", p.Name(), err)
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
						select {
						case shows <- s:
						case <-ctx.Done():
							return
						}
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
