package gulli

import (
	"net/url"
	"path/filepath"
	"strings"

	"github.com/gocolly/colly"

	"github.com/simulot/aspiratv/providers"
)

const (
	// http://www.gulli.fr/recherche?replaySearch[searchText]=il+%C3%A9tait+une+fois+l%27homme&replaySearch%5BsearchFilter%5D=videos
	searchURL = "http://www.gulli.fr/recherche?replaySearch[searchText]=%s&replaySearch[searchFilter]=videos"
)

func (p *Gulli) searchAll(mm []*providers.MatchRequest) ([]*providers.Show, error) {
	shows := []*providers.Show{}
	for _, m := range mm {
		if m.Provider == p.Name() {
			ID, showURL, err := p.searchPlayer(m)
			if err != nil {
				return nil, err
			}

			s, err := p.getPlayer(showURL, ID, m.Destination)
			if err != err {
				return nil, err
			}
			shows = append(shows, s...)
		}
	}
	return shows, nil
}

func (p *Gulli) searchPlayer(m *providers.MatchRequest) (ID string, showURL string, err error) {
	u := url.URL{
		Scheme: "http",
		Host:   "www.gulli.fr",
		Path:   "recherche",
	}
	q := u.Query()
	q.Set("replaySearch[searchText]", m.Show)
	q.Set("replaySearch[searchFilter]", "videos")
	u.RawQuery = q.Encode()

	done := false
	ID = ""
	parser := p.htmlParserFactory.New()
	parser.OnHTML("div.search-result>ul>li>div.titre", func(e *colly.HTMLElement) {
		if done {
			return
		}
		show := e.ChildText("a")
		if len(show) > 0 {
			if strings.Contains(strings.ToLower(show), m.Show) {
				showURL = "http:" + e.ChildAttr("a", "href")
				ID = filepath.Base(showURL)
				done = true
			}
		}
	})
	err = parser.Visit(u.String())
	if err != nil {
		ID = ""
		showURL = ""
	}
	return
}
