package gulli

import (
	"html"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/simulot/aspiratv/providers"
)

const gullyPlayer = "http://replay.gulli.fr/jwplayer/embed/" // + VOD ID

var reTitle = regexp.MustCompile(`^(.*)\s\-\sSaison\s(\d+),\sEpisode\s(\d+)\s:\s(.*)$`)
var reVars = regexp.MustCompile(
	`(?m)(?P<sources>sources:)` +
		`|(?:file:\s*(?U:"(?P<file>[^"]*)"))` +
		`|(?:mediaid:\s*(?U:"(?P<mediaid>[^"]*)"))` +
		`|(?:playlist_title:\s*(?U:"(?P<playlist_title>[^"]*)"))` +
		`|(?:image:\s*(?U:"(?P<image>[^"]*)"))`)

func (p *Gulli) getPlayer(ShowUrl, ID string, destination string) ([]*providers.Show, error) {

	r, err := p.getter.Get(gullyPlayer + ID)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	BaseShowURL := ShowUrl
	i := strings.LastIndex(ShowUrl, "/")
	if i >= 0 {
		BaseShowURL = ShowUrl[:i+1]
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
						show.Show = html.UnescapeString(t[0][1])
						show.Season = t[0][2]
						show.Episode = t[0][3]
						show.Title = html.UnescapeString(t[0][4])
					}
				case "mediaid":
					show.ID = s
					show.ShowURL = BaseShowURL + show.ID
					show.Provider = p.Name()
					show.Destination = destination
				}
			}
		}
	}
	if show != nil {
		shows = append(shows, show)
		p.seenShows[show.ID] = true
	}

	return shows, err
}
