package gulli

import (
	"html"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/simulot/aspiratv/providers"
)

const gullyPlayer = "http://replay.gulli.fr/jwplayer/embed/" // + VOD ID

var reTitle = regexp.MustCompile(`^(?P<show>.*)\s-\sS(?P<saison>\d+)\sép.\s(?P<episode>\d+)\s+:\s(?P<title>.*)$`)
var reVars = regexp.MustCompile(
	`(?m)(?P<sources>sources:)` +
		`|(?:file:\s*(?U:"(?P<file>[^"]*)"))` +
		`|(?:mediaid:\s*(?U:"(?P<mediaid>[^"]*)"))` +
		`|(?:playlist_title:\s*(?U:"(?P<playlist_title>[^"]*)"))` +
		`|(?:image:\s*(?U:"(?P<image>[^"]*)"))` +
		`|(?:description:\s*(?U:"(?P<description>[^"]*)"))`)

func (p *Gulli) getPlayer(ShowURL, ID string, destination string) ([]*providers.Show, error) {

	r, err := p.getter.Get(gullyPlayer + ID)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	BaseShowURL := ShowURL
	i := strings.LastIndex(ShowURL, "/")
	if i >= 0 {
		BaseShowURL = ShowURL[:i+1]
	}

	match := reVars.FindAllStringSubmatch(string(b), -1)

	shows := []*providers.Show{}

	var show *providers.Show
	parts := reVars.SubexpNames()

	for _, m := range match {
		for i, s := range m {
			if i > 0 && len(s) > 0 {
				switch parts[i] {
				case "sources":
					if show != nil {
						if !p.seenShows[show.ID] {
							shows = append(shows, show)
						}
						p.seenShows[show.ID] = true
					}
					show = &providers.Show{}
				case "file":
					if strings.HasSuffix(strings.ToLower(s), ".m3u8") {
						show.StreamURL = s
					}
				case "image":
					show.ThumbnailURL = s
				case "playlist_title":
					t := reTitle.FindAllStringSubmatch(s, -1)
					if len(t) > 0 && len(t[0]) == 5 {
						p2 := reTitle.SubexpNames()
						for j, s2 := range t[0] {
							switch p2[j] {
							case "show":
								show.Show = html.UnescapeString(s2)
							case "saison":
								show.Season = s2
							case "episode":
								show.Episode = s2
							case "title":
								show.Title = html.UnescapeString(s2)
							}
						}
					}
				case "mediaid":
					show.ID = s
					show.ShowURL = BaseShowURL + show.ID
					show.Provider = p.Name()
					show.Destination = destination
					show.Channel = "Gulli"
				case "description":
					show.Pitch = html.UnescapeString(s)
				}
			}
		}
	}
	if show != nil {
		if _, ok := p.seenShows[show.ID]; !ok {
			shows = append(shows, show)
			p.seenShows[show.ID] = true
		}
	}

	return shows, err
}
