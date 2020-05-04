package gulli

import (
	"context"
	"strings"

	"github.com/gocolly/colly"
)

type ShowEntry struct {
	Title    string
	URL      string
	ThumbURL string
}

var catalogURL = "https://replay.gulli.fr/"

func (p *Gulli) downloadCatalog(ctx context.Context) ([]ShowEntry, error) {
	ctx, done := context.WithTimeout(ctx, p.deadline)
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

	p.config.Log.Debug().Printf("[%s] Catalog url: %q", p.Name(), catalogURL)
	err := parser.Visit(catalogURL)
	if err != nil {
		return nil, err
	}
	return cat, nil
}
