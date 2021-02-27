package francetv

import (
	"context"
	"sync"
	"time"

	"github.com/simulot/aspiratv/matcher"
	"github.com/simulot/aspiratv/media"
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
)

// FranceTV structure handles france-tv catalog of shows
type FranceTV struct {
	seasons sync.Map
	shows   sync.Map
	config  providers.ProviderConfig
}

// New setup a Show provider for France Télévisions
func New() (*FranceTV, error) {
	p := &FranceTV{}
	return p, nil
}

// Name return the name of the provider
func (FranceTV) Name() string { return "francetv" }

func (p *FranceTV) Configure(fns ...providers.ProviderConfigFn) {
	c := p.config
	for _, f := range fns {
		c = f(c)
	}
	p.config = c
}

// MediaList return media that match with matching list.
func (p *FranceTV) MediaList(ctx context.Context, mm []*matcher.MatchRequest) chan *media.Media {
	shows := make(chan *media.Media)

	go func() {
		defer close(shows)
		for _, m := range mm {
			p.config.Log.Trace().Printf("[%s] Check matching request for %q", p.Name(), m.Show)

			if m.Provider != "francetv" {
				continue
			}
			for s := range p.search(ctx, m) {
				select {
				case <-ctx.Done():
				default:
					shows <- s
				}
			}
		}
		p.config.Log.Trace().Printf("[%s] MediaList is done", p.Name())
	}()
	return shows
}

type player struct {
	Video struct {
		URL   string `json:"url"`
		Token string `json:"token"`
	} `json:video`
	Meta struct {
		ID              string    `json:"id"`
		PlurimediaID    string    `json:"plurimedia_id"`
		Title           string    `json:"title"`
		AdditionalTitle string    `json:"additional_title"`
		PreTitle        string    `json:"pre_title"`
		BroadcastedAt   time.Time `json:"broadcasted_at"`
		ImageURL        string    `json:"image_url"`
	} `json:"meta"`
}
