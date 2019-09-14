package gulli

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

type ShowEntry struct {
	Title string
	URL   string
}

var catalogURL = "https://replay.gulli.fr/"

func (p *Gulli) downloadCatalog(ctx context.Context) ([]ShowEntry, error) {
	ctx, done := context.WithTimeout(ctx, 30*time.Second)
	cat := []ShowEntry{}
	defer done()

	parser := p.htmlParserFactory.New()

	parser.OnHTML("div.bloc_linkmore ul.multicolumn>li>a", func(e *colly.HTMLElement) {

		entry := ShowEntry{
			Title: strings.TrimSpace(e.Text),
			URL:   e.Attr("href"),
		}
		cat = append(cat, entry)
	})

	if p.debug {
		log.Println("[%s] Catalog url: %q", p.Name(), catalogURL)
	}
	err := parser.Visit(catalogURL)
	if err != nil {
		return nil, err
	}
	return cat, nil
}
