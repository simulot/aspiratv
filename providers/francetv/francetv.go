package francetv

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

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
			r := models.SearchResult{
				ID:       hit.ObjectID,
				Show:     firstStringOf(hit.Program.Label, hit.Label),
				Title:    hit.Title,
				PageURL:  home + "/" + firstStringOf(hit.URLComplete, hit.Path),
				ThumbURL: home + hit.Image.Formats["vignette_16x9"].Urls["w:1024"], //TODO: some images are missing
				Chanel:   firstStringOf(description.Channels[hit.Channel].Code, "francetv"),
				Plot:     firstStringOf(hit.Synopsis, hit.Description),
				Provider: "francetv",
				IsTeaser: hit.Type == "extrait",
				Season:   hit.SeasonNumber,
				Episode:  hit.EpisodeNumber,
				Aired:    hit.Dates["broadcast_begin_date"].Time(),
			}

			switch hit.Class {
			case "collection":
				r.Type = models.TypeCollection
			case "video":
				r.Type = models.TypeMovie
			case "program":
				r.Type = models.TypeSeries
			}

			for _, c := range hit.Categories {
				r.AddTag(c.Label)
			}

			if !q.OnlyExactTitle || q.IsMatch(r.Title) {
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
