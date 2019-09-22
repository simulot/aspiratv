package francetv

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
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

type Player struct {
	Video struct {
		URL   string `json:"url"`
		Token string `json:"token"`
	} `json:video`
}

// GetShowStreamURL return the show's URL, a m3u8 playlist
func (p *FranceTV) GetShowStreamURL(ctx context.Context, s *providers.Show) (string, error) {
	if s.StreamURL == "" {
		v := url.Values{}
		v.Set("country_code", "FR")
		v.Set("w", "1920")
		v.Set("h", "1080")
		v.Set("version", "5.18.3")
		v.Set("domain", "www.france.tv")
		v.Set("device_type", "desktop")
		v.Set("browser", "firefox")
		v.Set("browser_version", "69")
		v.Set("os", "windows")
		v.Set("gmt", "+1")

		u := "https://player.webservices.francetelevisions.fr/v1/videos/" + s.ID + "?" + v.Encode()

		if p.debug {
			log.Printf("[%s] Player url %q", p.Name(), u)
		}

		r, err := p.getter.Get(ctx, u)
		if err != nil {
			return "", fmt.Errorf("Can't get player: %w", err)
		}
		if p.debug {
			r = httptest.DumpReaderToFile(r, "francetv-player-"+s.ID+"-")
		}
		defer r.Close()

		pl := Player{}
		err = json.NewDecoder(r).Decode(&pl)
		if err != nil {
			return "", fmt.Errorf("Can't decode player: %w", err)
		}

		s.StreamURL = pl.Video.URL

		// Get Token
		if len(pl.Video.Token) > 0 {
			if p.debug {
				log.Printf("[%s] Player token %q", p.Name(), pl.Video.Token)
			}

			r2, err := p.getter.Get(ctx, pl.Video.Token)
			if err != nil {
				return "", fmt.Errorf("Can't get token %s: %w", pl.Video.Token, err)
			}
			if p.debug {
				r2 = httptest.DumpReaderToFile(r2, "francetv-token-"+s.ID+"-")
			}
			defer r2.Close()
			pl := struct {
				URL string `json:"url"`
			}{}
			err = json.NewDecoder(r2).Decode(&pl)
			if err != nil {
				return "", fmt.Errorf("Can't decode token: %w", err)
			}
			s.StreamURL = pl.URL
		}

		if p.debug {
			log.Printf("[%s] Stream url %q", p.Name(), s.StreamURL)
		}

	}
	return s.StreamURL, nil
}

// GetShowInfo query the URL from InfoOeuvre web service
func (p *FranceTV) GetShowInfo(ctx context.Context, s *providers.Show) error {
	return nil
}
