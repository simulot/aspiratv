package gulli

import (
	"github.com/gocolly/colly"
)

type gulliShow struct {
	name      string
	url       string
	thumbnail string
}

func getAllShowList(parser *colly.Collector, url string) ([]*gulliShow, error) {
	gss := []*gulliShow{}

	parser.OnHTML("div .program", func(e *colly.HTMLElement) {
		gs := &gulliShow{}
		if s := e.DOM.Find("a"); s != nil {
			if s, ok := s.Attr("href"); ok {
				gs.url = s
			}
		}
		if s := e.DOM.Find("img"); s != nil {
			if s, ok := s.Attr("src"); ok {
				gs.thumbnail = s
			}
			if s, ok := s.Attr("alt"); ok {
				gs.name = s
			}
		}
		gss = append(gss, gs)
	})

	err := parser.Visit(url)

	return gss, err
}
