package francetv

import (
	"context"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/simulot/aspiratv/matcher"
	"github.com/simulot/aspiratv/media"
	"github.com/simulot/aspiratv/net/myhttp"
	"github.com/simulot/aspiratv/parsers/htmlparser"
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

type getter interface {
	Get(ctx context.Context, uri string) (io.ReadCloser, error)
	DoWithContext(ctx context.Context, method string, theURL string, headers http.Header, body io.Reader) (io.ReadCloser, error)
}

// FranceTV structure handles france-tv catalog of shows
type FranceTV struct {
	getter            getter
	htmlParserFactory *htmlparser.Factory
	deadline          time.Duration
	seasons           sync.Map
	shows             sync.Map
	config            providers.Config
}

// WithGetter inject a getter in FranceTV object instead of normal one
func WithGetter(g getter) func(ftv *FranceTV) {
	return func(ftv *FranceTV) {
		ftv.getter = g
	}
}

// New setup a Show provider for France Télévisions
func New() (*FranceTV, error) {
	p := &FranceTV{
		getter:            myhttp.DefaultClient,
		deadline:          30 * time.Second,
		htmlParserFactory: htmlparser.NewFactory(),
	}
	return p, nil
}

// Name return the name of the provider
func (FranceTV) Name() string { return "francetv" }

func (p *FranceTV) Configure(c providers.Config) {
	p.config = c
	if p.config.Log.IsDebug() {
		p.deadline = 1 * time.Hour
	}
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
				shows <- s
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
