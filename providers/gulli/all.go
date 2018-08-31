package gulli

import (
	"log"
	"path/filepath"
	"strconv"

	"github.com/gocolly/colly"

	"github.com/simulot/aspiratv/providers"
)

const (
	gulliAllURL  = "http://replay.gulli.fr/all/always" // + page
	gulliWeekURL = "http://replay.gulli.fr/all/week"   // + page
)

var runCounter = 0

func (p *Gulli) getAll(mm []*providers.MatchRequest) ([]*providers.Show, error) {
	url := gulliAllURL
	if runCounter > 0 {
		url = gulliWeekURL
	}
	runCounter++

	shows := []*providers.Show{}
	done := false
	page := 0
	for !done {
		done = true
		parser := p.htmlParserFactory.New()

		parser.OnHTML("ul#infinite_scroll>li", func(e *colly.HTMLElement) {
			done = false

			show := &providers.Show{}
			show.ShowURL = e.ChildAttr("a", "href")
			show.ID = filepath.Base(show.ShowURL)
			show.ThumbnailURL = e.ChildAttr("img", "src")
			show.Show = e.ChildText("span.title")
			show.Title = e.ChildText("span.episode_title")
			show.Provider = p.Name()
			if p.debug {
				log.Println("[gulli] Seen ", show.Show, show.Title)
			}
			if p.seenShows[show.ID] {
				return // Skip if we have seen it already
			}
			if providers.IsShowMatch(mm, show) {
				s, err := p.getPlayer(show.ShowURL, show.ID, show.Destination)
				if err != nil {
					log.Println("[gulli] Can't load player at ", show.ShowURL)
				} else {
					shows = append(shows, s...)
				}
			}
		})
		u := url
		if page > 0 {
			u = url + "/" + strconv.Itoa(page)
		}
		if p.debug {
			log.Println("[gulli] Fetch page ", u)
		}
		err := parser.Visit(u)
		if err != nil {
			return nil, err
		}
		page++
	}
	return shows, nil
}
