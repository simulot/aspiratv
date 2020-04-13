package francetv

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"sync"
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
	getter      getter
	debug       bool
	deadline    time.Duration
	algolia     *AlgoliaConfig
	seasons     sync.Map
	shows       sync.Map
	keepBonuses bool
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
		getter:      myhttp.DefaultClient,
		deadline:    30 * time.Second,
		keepBonuses: true,
	}

	return p, nil
}

// Name return the name of the provider
func (FranceTV) Name() string { return "francetv" }

func (p *FranceTV) Configure(c providers.Config) {
	p.keepBonuses = c.KeepBonus
	p.debug = c.Debug
	if p.debug {
		p.deadline = time.Hour
	} else {
		p.deadline = 30 * time.Second
	}

}

// MediaList return media that match with matching list.
func (p *FranceTV) MediaList(ctx context.Context, mm []*providers.MatchRequest) chan *providers.Media {
	err := p.getAlgoliaConfig(ctx)

	if err != nil {
		return nil
	}
	shows := make(chan *providers.Media)

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

// GetMediaDetails download more details when available
func (p *FranceTV) GetMediaDetails(ctx context.Context, m *providers.Media) error {
	info := m.Metadata.GetMediaInfo()
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

	u := "https://player.webservices.francetelevisions.fr/v1/videos/" + m.ID + "?" + v.Encode()

	if p.debug {
		log.Printf("[%s] Player url %q", p.Name(), u)
	}

	r, err := p.getter.Get(ctx, u)
	if err != nil {
		return fmt.Errorf("Can't get player: %w", err)
	}
	if p.debug {
		r = httptest.DumpReaderToFile(r, "francetv-player-"+m.ID+"-")
	}
	defer r.Close()

	pl := player{}
	err = json.NewDecoder(r).Decode(&pl)
	if err != nil {
		return fmt.Errorf("Can't decode player: %w", err)
	}

	episodeRegexp := regexp.MustCompile(`S(\d+)\sE(\d+)`)
	expr := episodeRegexp.FindAllStringSubmatch(pl.Meta.PreTitle, -1)
	if len(expr) > 0 {
		info.Season, _ = strconv.Atoi(expr[0][1])
		info.Episode, _ = strconv.Atoi(expr[0][2])
	}

	// Get Token
	if len(pl.Video.Token) > 0 {
		if p.debug {
			log.Printf("[%s] Player token %q", p.Name(), pl.Video.Token)
		}

		r2, err := p.getter.Get(ctx, pl.Video.Token)
		if err != nil {
			return fmt.Errorf("Can't get token %s: %w", pl.Video.Token, err)
		}
		if p.debug {
			r2 = httptest.DumpReaderToFile(r2, "francetv-token-"+m.ID+"-")
		}
		defer r2.Close()
		pl := struct {
			URL string `json:"url"`
		}{}
		err = json.NewDecoder(r2).Decode(&pl)
		if err != nil {
			return fmt.Errorf("Can't decode token's url : %w", err)
		}
		if p.debug {
			log.Printf("[%s] Player token's url %q", p.Name(), pl.URL)
		}

		// Now, get pl.URL, and watch for Location response header. It contains the dash ressource
		// Set up the HTTP request
		req, err := http.NewRequest("GET", pl.URL, nil)
		if err != nil {
			return err
		}

		transport := http.Transport{}
		resp, err := transport.RoundTrip(req)
		if err != nil {
			return err
		}

		// Check if you received the status codes you expect. There may
		// status codes other than 200 which are acceptable.
		if resp.StatusCode != 302 {
			return fmt.Errorf("Failed with status: %q", resp.Status)
		}

		info.URL = resp.Header.Get("Location")

	}

	if p.debug {
		log.Printf("[%s] Stream url %q", p.Name(), info.URL)
	}

	return nil
}
