package francetv

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/simulot/aspiratv/models"
	"github.com/simulot/aspiratv/myhttp"
	"github.com/simulot/aspiratv/providers"
	"golang.org/x/time/rate"
)

const (
	home = "https://www.france.tv"
)

var (
	description = providers.Description{
		Code: "francetv",
		Name: "france•tv",
		Logo: "/web/francetv.png",
		URL:  "https://www.france.tv/",
		Channels: map[string]providers.Channel{
			"francetv": {
				Code: "francetv",
				Name: "france•tv",
				Logo: "/web/francetv.png",
				URL:  "https://www.france.tv/",
			},
			"france-2": {
				Code: "france-2",
				Name: "France 2",
				Logo: "/web/france-2.svg",
				URL:  "https://www.france.tv/france-2/",
			},
			"france-3": {
				Code: "france-3",
				Name: "France 3",
				Logo: "/web/france-3.svg",
				URL:  "https://www.france.tv/france-3/",
			},
			"france-4": {
				Code: "france-4",
				Name: "France 4",
				Logo: "/web/france-4.svg",
				URL:  "https://www.france.tv/france-4/",
			},
			"france-5": {
				Code: "france-5",
				Name: "France 5",
				Logo: "/web/france-5.svg",
				URL:  "https://www.france.tv/france-5/",
			},
			"slash": {
				Code: "slash",
				Name: "France tv Slash",
				Logo: "/web/slash.svg",
				URL:  "https://www.france.tv/slash/",
			},
			"la1ere": {
				Code: "la1ere",
				Name: "la 1ère ",
				Logo: "/web/la1ere.svg",
				URL:  "https://www.france.tv/la1ere/",
			},
			"franceinfo": {
				Code: "franceinfo",
				Name: "franceinfo",
				Logo: "/web/franceinfo.svg",
				URL:  "https://www.france.tv/franceinfo/",
			},
		},
	}
)

type FranceTV struct {
	client  *myhttp.Client
	limiter *rate.Limiter
}

func WithClientConfigurations(confFn ...func(c *myhttp.Client)) func(p *FranceTV) {
	return func(p *FranceTV) {
		p.client = myhttp.NewClient(confFn...)
	}
}

func WithLimiter(l *rate.Limiter) func(p *FranceTV) {
	return func(p *FranceTV) {
		p.limiter = l
	}
}

func NewFranceTV(confFn ...func(p *FranceTV)) *FranceTV {
	p := FranceTV{}

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

func (FranceTV) ProviderDescribe(ctx context.Context) providers.Description {
	return description
}

func (p *FranceTV) Search(ctx context.Context, q models.SearchQuery) (<-chan models.SearchResult, error) {
	results := make(chan models.SearchResult, 1)

	go p.callSearch(ctx, results, q)
	return results, nil
}

func (p *FranceTV) callSearch(ctx context.Context, results chan models.SearchResult, q models.SearchQuery) {
	defer close(results)

	query := RequestPayLoad{
		Term: q.Title,
		Options: &Options{
			ContentsLimit: 20,
			Types:         "content",
		},
	}

	req, err := p.client.NewRequestJSON(ctx, home+"/recherche/lancer/", nil, nil, query)
	if err != nil {
		results <- models.SearchResult{Err: err}
		return
	}

	var response string
	p.limiter.Wait(ctx)
	if err != nil {
		results <- models.SearchResult{Err: err}
		return
	}

	select {
	case <-ctx.Done():
		return
	default:
		err = p.client.PostJSON(req, &response)
		if err != nil {
			results <- models.SearchResult{Err: err}
			return
		}
	}

	// TODO: remove this
	// dump the JSON structure
	f, _ := os.Create("dump_francetv.json")
	f.WriteString(response)
	f.Close()

	// Must decode the response another time
	founds := map[string]Result{}
	err = json.Unmarshal([]byte(response), &founds)
	if err != nil {
		log.Print(err)
		results <- models.SearchResult{Err: err}
		return
	}

	rank := 0

	for _, cat := range []string{"taxonomy", "collection", "content"} {
		for _, hit := range founds[cat].Hits {

			if hit.Type == "extrait" {
				continue
			}

			u := firstStringOf(hit.URLComplete, hit.Path)
			if u == "" {
				continue
			}

			r := models.SearchResult{
				ID:       hit.ObjectID,
				Show:     firstStringOf(hit.Program.Label, hit.Label),
				Title:    hit.Title,
				PageURL:  home + "/" + u,
				ThumbURL: home + hit.Image.Formats["vignette_16x9"].Urls["w:1024"], //TODO: some images are missing
				Chanel:   firstStringOf(description.Channels[hit.Channel].Code, "francetv"),
				Plot:     firstStringOf(hit.Synopsis, hit.Text, hit.Description),
				Provider: "francetv",
				Aired:    hit.Dates["broadcast_begin_date"].Time(),
			}

			switch hit.Class {
			case "collection":
				r.Type = models.TypeCollection
			case "video":
				r.Type = models.TypeMovie
				continue
			case "program":
				r.Type = models.TypeSeries
			}

			for _, c := range hit.Categories {
				r.AddTag(c.Label)
			}

			if r.Type != models.TypeMovie && (!q.OnlyExactTitle || q.IsMatch(r.Title)) {
				err := p.GetCollectionDetails(ctx, hit, &r, q)
				if err != nil {
					log.Printf("[ARTE] Can't get collection details: %s", err)
					continue
				}
				if r.AvailableVideos == 0 {
					continue
				}
				rank++
				r.Rank = rank
				select {
				case <-ctx.Done():
					return
				case results <- r:
				}
			}
		}

	}
}

func (p *FranceTV) GetCollectionDetails(ctx context.Context, hit Hits, r *models.SearchResult, q models.SearchQuery) error {

	page := 0
	r.Type = models.TypeCollection
	for {
		hits := 0
		u := fmt.Sprintf("%s/toutes-les-videos/?page=%d", r.PageURL, page)
		err := p.limiter.Wait(ctx)
		if err != nil {
			return err
		}
		parser := colly.NewCollector()
		parser.OnHTML("a.c-card-video", func(e *colly.HTMLElement) {
			if strings.Contains(e.Attr("class"), "c-card-video--unavailable") {
				return
			}
			var (
				extract bool
				aired   time.Time
			)
			e.ForEach("span", func(i int, e *colly.HTMLElement) {
				if e.Text == "extrait" {
					extract = true
					return
				}
				cl := strings.Split(e.Attr("class"), " ")
				for _, c := range cl {
					switch c {
					case "c-card-video__textarea-title":
						if r.Show != e.Text {
							r.Type = models.TypeCollection
						} else {
							r.Type = models.TypeTVShow
						}
					case "c-card-video__textarea-subtitle":
						if match := reAnalyseTitle.FindStringSubmatch(e.Text); len(match) == 4 {
							r.Type = models.TypeSeries
						}
					case "c-label":
						extract = true
					case "c-metadata":
						if match := reAired.FindStringSubmatch(e.Text); len(match) == 3 {
							day, _ := strconv.Atoi(match[1])
							month, _ := strconv.Atoi(match[2])
							year := time.Now().Year()
							d := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
							if d.After(time.Now()) {
								d = time.Date(year-1, time.Month(month), day, 0, 0, 0, 0, time.Local)
							}
							aired = d
							if aired.Before(q.AiredAfter) {
								return
							}
						}
					}
				}
			})
			if extract {
				return
			}
			r.AvailableVideos++
			hits++

		})
		log.Printf("[HTTPCLIENT] GET %s", u)
		err = parser.Visit(u)
		if err != nil {
			return err
		}
		page++
		if page == 2 && hits > 0 {
			r.MoreAvailable = true
			return nil
		}
		if hits == 0 {
			return nil
		}
	}
}

var reAnalyseTitle = regexp.MustCompile(`^\s?S(\d+)?\s+E(\d+)\s+-\s+(.*)$`)
var reAired = regexp.MustCompile(`(\d{2})\/(\d{2})`)

func joinStrings(sep string, ss ...string) string {
	r := strings.Builder{}

	for i, s := range ss {
		if len(s) > 0 {
			if i > 0 && ss[i-1] == s {
				continue
			}
			if r.Len() > 0 {
				r.WriteString(sep)
			}
			r.WriteString(s)
		}
	}
	return r.String()
}

func firstStringOf(ss ...string) string {
	for _, s := range ss {
		if len(s) > 0 {
			return s
		}
	}
	return ""
}
