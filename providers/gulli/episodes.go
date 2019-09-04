package gulli

import (
	"path"

	"github.com/gocolly/colly"
)

func (p *Gulli) getFirstEpisodeID(entry ShowEntry) (string, error) {

	parser := p.htmlParserFactory.New()
	var playerURL string

	parser.OnHTML("div.bloc.bloc_listing ul li:first-child a", func(e *colly.HTMLElement) {
		playerURL = e.Attr("href")
	})
	err := parser.Visit(entry.URL)
	if err != nil {
		return "", err
	}
	ID := path.Base(playerURL)
	return ID, nil
}
