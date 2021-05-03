package artetv

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/simulot/aspiratv/pkg/models"
	"github.com/simulot/aspiratv/pkg/providers"

	"github.com/simulot/aspiratv/pkg/myhttp"
	"golang.org/x/time/rate"
)

type Arte struct {
	preferredLanguage string
	client            *myhttp.Client
	limiter           *rate.Limiter
}

var (
	channels = providers.Description{
		Code: "arte",
		Name: "arte",
		Logo: "/web/arte.png",
		Channels: map[string]providers.Channel{
			"arte": {
				Code: "arte",
				Name: "arte",
				Logo: "/web/arte.png",
			},
		},
	}
)

func WithClientConfigurations(confFn ...func(c *myhttp.Client)) func(p *Arte) {
	return func(p *Arte) {
		p.client = myhttp.NewClient(confFn...)
	}
}
func WithLimiter(l *rate.Limiter) func(p *Arte) {
	return func(p *Arte) {
		p.limiter = l
	}
}

func NewArte(confFn ...func(p *Arte)) *Arte {
	p := Arte{
		preferredLanguage: "fr",
	}
	for _, fn := range confFn {
		fn(&p)
	}
	if p.client == nil {
		p.client = myhttp.NewClient()
	}
	if p.limiter == nil {
		p.limiter = rate.NewLimiter(rate.Every(300*time.Millisecond), 5)
	}

	return &p
}

func (Arte) Name() string { return "arte" }

func (Arte) ProviderDescribe(ctx context.Context) providers.Description {
	return channels
}

// Search query hte web site and reports items that are matching the query.
// Series and Shows are listed as a whole, with MediaType Show. The actual type is unknown
func (p *Arte) Search(ctx context.Context, q models.SearchQuery) (<-chan models.SearchResult, error) {
	results := make(chan models.SearchResult, 1)

	go p.callSearch(ctx, results, q)
	return results, nil
}

func (p *Arte) callSearch(ctx context.Context, results chan models.SearchResult, q models.SearchQuery) {
	defer close(results)
	var guide GuideResult

	select {
	case <-ctx.Done():
		results <- models.SearchResult{
			Err: ctx.Err(),
		}
		return
	default:

		urlValues := url.Values{}
		urlValues.Add("query", q.Title)
		urlValues.Add("mainZonePage", "1")
		req, err := p.client.NewRequest(
			ctx,
			"https://www.arte.tv/guide/api/emac/v3/%s/web/pages/SEARCH/", []interface{}{p.preferredLanguage},
			urlValues,
			nil,
		)
		if err != nil {
			results <- models.SearchResult{
				Err: err,
			}
			return
		}

		err = p.limiter.Wait(ctx)
		if err != nil {
			results <- models.SearchResult{
				Err: ctx.Err(),
			}
			return
		}

		err = p.client.GetJSON(req, &guide)
		if err != nil {
			results <- models.SearchResult{
				Err: ctx.Err(),
			}
			return
		}

		rank := 0
		for _, z := range guide.Zone {
			if z.Code.Name != "listing_SEARCH" {
				continue
			}
			for _, d := range z.Data {

				mediaType := models.TypeUnknown
				if d.Kind.IsCollection {
					mediaType = models.TypeCollection
				} else {
					mediaType = models.TypeMovie
					continue
				}

				r := models.SearchResult{
					Title:       joinStrings(" - ", d.Title, d.Subtitle),
					Plot:        firstStrings(d.TeaserText, d.ShortDescription),
					PageURL:     d.URL,
					ID:          d.ID,
					ThumbURL:    bestImage(d.Images["landscape"]),
					Type:        mediaType,
					Chanel:      "arte",
					Provider:    "arte",
					AvailableOn: d.Availability.Start.Time(),
				}

				if !q.OnlyExactTitle || q.IsMatch(r.Title) {
					rank++
					r.Rank = rank
					// Series are not detailled
					if d.Kind.IsCollection {
						err := p.GetCollectionDetails(ctx, d, &r)
						if err != nil {
							log.Printf("[ARTE] Can't get collection details: %s", err)
							continue
						}
					}
					select {
					case <-ctx.Done():
						return
					case results <- r:
					}
				}
			}
		}
	}
}

func (p *Arte) GetCollectionDetails(ctx context.Context, d Data, r *models.SearchResult) error {
	req, err := p.client.NewRequest(ctx, d.URL, nil, nil, nil)
	if err != nil {
		return err
	}

	err = p.limiter.Wait(ctx)
	if err != nil {
		return err
	}

	var resp *http.Response
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		resp, err = p.client.Get(req)
	}
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	b, err = extractState(b, "__INITIAL_STATE__")
	if err != nil {
		return err
	}

	state := InitialProgram{}
	err = json.Unmarshal(b, &state)
	if err != nil {
		return err
	}
	for _, page := range state.Pages.List {
		for _, zone := range page.Zones {
			if zone.Code.Name == "collection_content" || zone.Code.Name == "program_content" {
				for _, d := range zone.Data {
					switch d.Kind.Code {
					case "TV_SERIES":
						r.Type = models.TypeSeries
					case "MAGAZINE":
						r.Type = models.TypeTVShow
					default:
						r.Type = models.TypeCollection
					}
					r.Show = d.Title
					r.Plot = d.ShortDescription
				}
			}
		}
	}
	return nil
}

func extractState(b []byte, tag string) ([]byte, error) {
	start := bytes.Index(b, []byte(tag))
	if start < 0 {
		return nil, fmt.Errorf("can't locate %q", tag)
	}
	b = b[start:]
	start = bytes.Index(b, []byte("{"))
	if start < 0 {
		return nil, fmt.Errorf("can't locate %q", tag)
	}
	end := bytes.Index(b, []byte("</script>"))
	if end < 0 {
		return nil, fmt.Errorf("can't locate %q", tag)
	}
	b = b[:end]

	end = bytes.LastIndex(b, []byte(";"))
	if end < 0 {
		return nil, fmt.Errorf("can't locate %q", tag)
	}
	b = bytes.TrimSuffix(b[start:], b[end:]) //TODO simplify this
	return b, nil
}

func joinStrings(sep string, ss ...string) string {
	r := strings.Builder{}

	for _, s := range ss {
		if len(s) > 0 {
			if r.Len() > 0 {
				r.WriteString(sep)
			}
			r.WriteString(s)
		}
	}
	return r.String()
}

func firstStrings(ss ...string) string {
	for _, s := range ss {
		if len(s) > 0 {
			return s
		}
	}
	return ""
}

func bestImage(images Image) string {
	bestURL := ""
	bestResolution := 0
	for _, r := range images.Resolutions {
		resolution := r.W * r.H
		if resolution > bestResolution {
			bestResolution = resolution
			bestURL = r.URL
		}
	}
	return bestURL
}
