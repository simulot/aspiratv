package gulli

import (
	"context"
	"log"
	"path"
	"time"

	"github.com/gocolly/colly"
)

func (p *Gulli) getFirstEpisodeID(ctx context.Context, entry ShowEntry) (string, error) {
	ctx, done := context.WithTimeout(ctx, 30*time.Second)
	defer done()
	parser := p.htmlParserFactory.New()
	var playerURL string

	parser.OnHTML("div.bloc.bloc_listing ul li:first-child a", func(e *colly.HTMLElement) {
		playerURL = e.Attr("href")
	})
	if p.debug {
		log.Println("[%s] Episode URL: %q", p.Name(), entry.URL)
	}
	err := parser.Visit(entry.URL)
	if err != nil {
		return "", err
	}
	ID := path.Base(playerURL)
	return ID, nil
}
