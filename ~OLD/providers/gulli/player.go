package gulli

import (
	"context"
	"html"
	"regexp"
	"strconv"
	"strings"

	"github.com/simulot/aspiratv/matcher"
	"github.com/simulot/aspiratv/media"
	"github.com/simulot/aspiratv/metadata/nfo"
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

func (p *Gulli) getPlayer(ctx context.Context, mr *matcher.MatchRequest, ID string) ([]*media.Media, error) {

	p.config.Log.Debug().Printf("[%s] Player URL: %q", p.Name(), gullyPlayer+ID)

	client := providers.NewHTTPClient(p.config)
	b, err := client.Get(ctx, gullyPlayer+ID, nil, nil)
	if err != nil {
		return nil, err
	}
	match := reVars.FindAllStringSubmatch(string(b), -1)

	shows := []*media.Media{}

	var info *nfo.MediaInfo

	parts := reVars.SubexpNames()
	for _, m := range match {
		for i, s := range m {
			if i > 0 && len(s) > 0 {
				switch parts[i] {
				case "sources":
					if info != nil {
						if !p.seenShows[info.UniqueID[0].ID] {
							shows = append(shows, &media.Media{
								ID:    info.UniqueID[0].ID,
								Match: mr,
								Metadata: &nfo.EpisodeDetails{
									MediaInfo: *info,
								},
							})
						}
						p.seenShows[info.UniqueID[0].ID] = true
					}
					info = &nfo.MediaInfo{
						MediaType: nfo.TypeSeries,
					}
				case "file":
					if strings.HasSuffix(strings.ToLower(s), ".m3u8") {
						info.MediaURL = s
					}
				case "image":
					info.Thumb = []nfo.Thumb{
						{
							Aspect: "thumb",
							URL:    s,
						},
					}
				case "playlist_title":
					t := reTitle.FindAllStringSubmatch(s, -1)
					if len(t) > 0 && len(t[0]) == 5 {
						p2 := reTitle.SubexpNames()
						for j, s2 := range t[0] {
							switch p2[j] {
							case "show":
								info.Showtitle = html.UnescapeString(s2)
							case "saison":
								info.Season, _ = strconv.Atoi(s2)
							case "episode":
								info.Episode, _ = strconv.Atoi(s2)
							case "title":
								info.Title = html.UnescapeString(s2)
							}
						}
					}
				case "mediaid":
					info.UniqueID = []nfo.ID{
						{
							ID:   s,
							Type: "GULLITV",
						},
					}
					info.Tag = []string{"Gulli"}
					info.Genre = []string{"dessins animés", "enfants"}
				case "description":
					info.Plot = html.UnescapeString(s)
				}
			}
		}
	}
	if info != nil && len(info.UniqueID) > 0 {
		if !p.seenShows[info.UniqueID[0].ID] {
			info.MediaType = nfo.TypeSeries
			if info.Episode == 0 && info.Season == 0 {
				info.MediaType = nfo.TypeMovie
			}
			shows = append(shows, &media.Media{
				ID:    info.UniqueID[0].ID,
				Match: mr,
				Metadata: &nfo.EpisodeDetails{
					MediaInfo: *info,
				},
			})
		}
		p.seenShows[info.UniqueID[0].ID] = true
	}
	return shows, err
}
