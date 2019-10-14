package gulli

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/gocolly/colly"
	"github.com/simulot/aspiratv/metadata/nfo"
)

var protectCartoonList sync.RWMutex

func (p *Gulli) getCartoonPage(ctx context.Context, showTitle string) (string, string, error) {
	showTitle = strings.ToLower(showTitle)

	protectCartoonList.Lock()
	if p.cartoonList == nil {
		l, err := p.getCartoonList(ctx)
		if err != nil {
			return "", "", fmt.Errorf("Can't get cartoons' page: %s", err)
		}
		p.cartoonList = l
	}
	protectCartoonList.Unlock()
	protectCartoonList.RLock()
	defer protectCartoonList.RUnlock()
	for _, c := range p.cartoonList {
		if strings.Contains(strings.ToLower(c.Title), showTitle) {
			return c.URL, c.ThumbURL, nil
		}
	}
	log.Printf("[%s] Can't find cartoon %q", p.Name(), showTitle)
	return "", "", fmt.Errorf("Cartoon %q not found", showTitle)
}

func (p *Gulli) getCartoonList(ctx context.Context) ([]ShowEntry, error) {
	const cartoonURL = "https://www.gulli.fr/Chaine-TV/Dessins-animes"
	page := 1

	l := []ShowEntry{}

	parser := p.htmlParserFactory.New()

	parser.OnHTML("div.listing-home li", func(e *colly.HTMLElement) {
		thumb := e.ChildAttr("span>img", "data-src")
		title := strings.TrimSpace(e.ChildText("div>a"))
		url := e.ChildAttr("div>a", "href")
		l = append(l, ShowEntry{
			Title:    title,
			URL:      url,
			ThumbURL: thumb,
		})
	})

	for {
		url := cartoonURL
		if page > 1 {
			url += fmt.Sprintf("?page=%d", page)
		}
		err := parser.Visit(url)
		if err != nil {
			return nil, err
		}
		if page >= 5 {
			break
		}
		page++
	}

	return l, nil
}

func (p *Gulli) getShowInfo(ctx context.Context, showTitle string) (*nfo.TVShow, error) {
	u, t, err := p.getCartoonPage(ctx, showTitle)
	if err != nil {
		return nil, err
	}

	tvshow := &nfo.TVShow{}
	parser := p.htmlParserFactory.New()

	parser.OnHTML("div.dossier-entete-visuel>img", func(e *colly.HTMLElement) {
		tvshow.Thumb = []nfo.Thumb{
			{
				Aspect: "fanart",
				URL:    e.Attr("src"),
			},
			{
				Aspect: "poster",
				URL:    t,
			},
		}
	})

	parser.OnHTML("div.dossier-chapo", func(e *colly.HTMLElement) {
		tvshow.Plot = e.ChildText("p")
	})
	parser.OnHTML("ol.breadcrumb>li:last-child ", func(e *colly.HTMLElement) {
		tvshow.Title = strings.TrimSpace(e.Text)
	})

	err = parser.Visit(u)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("Can't grab cartoon info: %w", err)
	}
	return tvshow, nil
}
