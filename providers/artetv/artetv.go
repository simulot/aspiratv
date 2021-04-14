package artetv

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/simulot/aspiratv/models"
	"github.com/simulot/aspiratv/providers"

	"github.com/simulot/aspiratv/providers/client"
	"golang.org/x/time/rate"
)

type Arte struct {
	preferredLanguage string
	client            *client.Client
	limiter           *rate.Limiter
}

func NewArte() *Arte {
	p := Arte{
		preferredLanguage: "fr",
		client:            client.New(),
		limiter:           rate.NewLimiter(rate.Every(300*time.Millisecond), 5),
	}

	return &p
}

func (Arte) ProviderDescribe(ctx context.Context) providers.ProviderDescription {
	return providers.ProviderDescription{
		Name:     "artetv",
		Channels: []string{"Arte"},
	}
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

		for _, z := range guide.Zone {
			if z.Code.Name != "listing_SEARCH" {
				continue
			}
			for _, d := range z.Data {
				mediaType := models.MediaTypeUnknown
				if d.Kind.IsCollection {
					mediaType = models.MediaTypeCollection
				} else {
					mediaType = models.MediaTypeMovie
				}

				if true || q.IsMatch(d.Title) {
					result := models.SearchResult{
						Title:       joinStrings(" - ", d.Title, d.Subtitle),
						Plot:        firstStrings(d.TeaserText, d.ShortDescription),
						PageURL:     d.URL,
						ID:          d.ID,
						ThumbURL:    bestImage(d.Images["landscape"]),
						Type:        mediaType,
						Chanel:      "Arte",
						Provider:    "artetv",
						IsPlayable:  false,
						IsTeaser:    mediaType != models.MediaTypeCollection,
						AvailableOn: d.Availability.Start.Time(),
					}

					for _, s := range d.Stickers {
						switch s.Code {
						case "PLAYABLE":
							result.IsPlayable = true
						case "FULL_VIDEO":
							result.IsTeaser = false
						}
					}

					select {
					case <-ctx.Done():
						return
					case results <- result:
					}
				}
			}
		}
	}
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
