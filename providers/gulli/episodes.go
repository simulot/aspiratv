package gulli

import (
	"context"
	"path"

	"github.com/gocolly/colly"
)

func (p *Gulli) getFirstEpisodeID(ctx context.Context, entry ShowEntry) (string, error) {

	parser := colly.NewCollector()
	playerURL := ""
	parser.OnHTML("div.bloc.bloc_listing ul li:first-child a", func(e *colly.HTMLElement) {
		playerURL = e.Attr("href")
	})
	p.config.Log.Debug().Printf("[%s] Episode URL: %q", p.Name(), entry.URL)
	err := parser.Visit(entry.URL)
	if err != nil {
		return "", err
	}
	ID := path.Base(playerURL)
	return ID, nil
}
