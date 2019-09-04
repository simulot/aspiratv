package gulli

import (
	"strings"

	"github.com/gocolly/colly"
)

type ShowEntry struct {
	Title string
	URL   string
}

var catalogURL = "https://replay.gulli.fr/"

func (p *Gulli) downloadCatalog() ([]ShowEntry, error) {
	cat := []ShowEntry{}

	parser := p.htmlParserFactory.New()

	parser.OnHTML("div.bloc_linkmore ul.multicolumn>li>a", func(e *colly.HTMLElement) {

		entry := ShowEntry{
			Title: strings.TrimSpace(e.Text),
			URL:   e.Attr("href"),
		}
		cat = append(cat, entry)
	})

	err := parser.Visit(catalogURL)
	if err != nil {
		return nil, err
	}
	return cat, nil
}
