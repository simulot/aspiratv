package francetv

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	getter   getter
	deadline time.Duration
	algolia  *AlgoliaConfig
	seasons  sync.Map
	shows    sync.Map
	config   providers.Config
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
		getter:   myhttp.DefaultClient,
		deadline: 30 * time.Second,
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

// GetMediaDetails download more details when available  especially the stream URL.
// The player webservice returns some metadata and an URL named Token.
// The must been acquired right before the actual download. It has a limited validity
// In the structure returned by token URL, another URL is provided. The request is then redirected
// to the actual video stream. This url has also a limited validity.
//
// But for some reason FFMPEG doesn't follow the redirection. So, we have to get the final URL before
// calling FFMPEG // FranceTV provides a subtitle tracks that isn't decoded by FFMPEG.
// And FFMPEG doesn't get always  the best video resolution
//
// The video stream is in fact a MPD manifest files. We can edit this manifest for removing unwanted tracks.
//
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
	v.Set("browser_version", "75")
	v.Set("os", "windows")
	v.Set("gmt", "+1")

	u := "https://player.webservices.francetelevisions.fr/v1/videos/" + m.ID + "?" + v.Encode()
	p.config.Log.Debug().Printf("[%s] Player URL for title '%s' is %q.", p.Name(), m.Metadata.GetMediaInfo().Title, u)

	r, err := p.getter.Get(ctx, u)
	if err != nil {
		return fmt.Errorf("Can't get player: %w", err)
	}
	if p.config.Log.IsDebug() {
		r = httptest.DumpReaderToFile(p.config.Log, r, "francetv-player-"+m.ID+"-")
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
		p.config.Log.Debug().Printf("[%s] Player token for '%s' is %q ", p.Name(), m.Metadata.GetMediaInfo().Title, pl.Video.Token)

		r2, err := p.getter.Get(ctx, pl.Video.Token)
		if err != nil {
			return fmt.Errorf("Can't get token %s: %w", pl.Video.Token, err)
		}
		if p.config.Log.IsDebug() {
			r2 = httptest.DumpReaderToFile(p.config.Log, r2, "francetv-token-"+m.ID+"-")
		}
		defer r2.Close()
		pl := struct {
			URL string `json:"url"`
		}{}
		err = json.NewDecoder(r2).Decode(&pl)
		if err != nil {
			return fmt.Errorf("Can't decode token's url : %w", err)
		}
		info.URL = pl.URL

	}
	p.config.Log.Trace().Printf("[%s] Player URL for '%s' is %q ", p.Name(), m.Metadata.GetMediaInfo().Title, info.URL)
	return nil
}
