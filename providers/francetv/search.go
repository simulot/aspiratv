package francetv

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/simulot/aspiratv/metadata/nfo"
	"github.com/simulot/aspiratv/net/myhttp/httptest"
	"github.com/simulot/aspiratv/providers"
	"github.com/simulot/aspiratv/providers/francetv/query"
	"github.com/simulot/aspiratv/providers/matcher"
)

const homeFranceTV = "https://www.france.tv"

type RequestPayLoad struct {
	Term    string  `json:"term"`
	Signal  Signal  `json:"signal"`
	Options Options `json:"options"`
}
type Signal struct {
}
type Options struct {
	ContentsLimit   int    `json:"contentsLimit,omitempty"`
	TaxonomiesLimit int    `json:"taxonomiesLimit,omitempty"`
	Types           string `json:"types"`
}

func (p *FranceTV) search(ctx context.Context, mr *matcher.MatchRequest) chan *providers.Media {
	mm := make(chan *providers.Media)

	go func() {
		defer close(mm)
		// ctx, done := context.WithTimeout(ctx, p.deadline)
		// defer done()

		rq := RequestPayLoad{
			Term: mr.Show,
			Options: Options{
				ContentsLimit: 20,
				// TaxonomiesLimit: 20,
				Types: "content",
			},
		}
		encRq, err := json.Marshal(rq)
		if err != nil {
			p.config.Log.Error().Printf("[%s] Can't encode request: %s", p.Name(), err)
			return
		}

		h := make(http.Header)

		if p.config.Log.IsDebug() {
			p.config.Log.Debug().Printf("[%s] Request headers", p.Name())
			for k, s := range h {
				p.config.Log.Debug().Printf("[%s] %q %s", p.Name(), k, strings.Join(s, ","))
			}
			p.config.Log.Debug().Printf(string(encRq))
		}

		r, err := p.getter.DoWithContext(ctx, "POST", "https://www.france.tv/recherche/lancer/", h, bytes.NewBuffer(encRq))
		if err != nil {
			p.config.Log.Error().Printf("[%s] Can't call search API: %s", p.Name(), err)
			return
		}
		if p.config.Log.IsDebug() {
			r = httptest.DumpReaderToFile(p.config.Log, r, "francetv-recherche-")
		}

		resp, err := ioutil.ReadAll(r)
		r.Close()
		if err != nil {
			p.config.Log.Error().Printf("[%s] Can't get API result: %s", p.Name(), err)
			return
		}

		// resp is an encoded string containing a json object.
		decResp := ""
		err = json.Unmarshal(resp, &decResp)
		if err != nil {
			p.config.Log.Error().Printf("[%s] Can't decode search response: %s", p.Name(), err)
			return
		}

		results := map[string]query.Result{}
		err = json.Unmarshal([]byte(decResp), &results)
		if err != nil {
			p.config.Log.Error().Printf("[%s] Can't decode API result: %s", p.Name(), err)
			return
		}
		if p.config.Log.IsDebug() {
			p.config.Log.Trace().Printf("Decoded result")
			reEncode, err := json.MarshalIndent(results, "", "  ")
			if err != nil {
				p.config.Log.Error().Printf("Can't encode json response: %s", err)
			}
			p.config.Log.Trace().Printf("\n%s", string(reEncode))
		}
		for k, result := range results {
			_ = k
			for _, h := range result.Hits {
				if h.Type != "integrale" {
					continue
				}

				found := false
				found = found || strings.Contains(strings.ToLower(h.Program.Label), mr.Show)
				found = found || strings.Contains(strings.ToLower(h.Title), mr.Show)
				if !found {
					continue
				}

				media := &providers.Media{
					ID:    h.SiID.String(),
					Match: mr,
				}
				var info *nfo.MediaInfo

				if h.SeasonNumber != 0 || h.Program.Class == "program" {
					meta := nfo.EpisodeDetails{}
					info = &meta.MediaInfo
					media.SetMetaData(&meta)
					media.ShowType = providers.Series

				} else {
					meta := nfo.Movie{}
					info = &meta.MediaInfo
					media.SetMetaData(&meta)
					media.ShowType = providers.Movie
				}

				*info = nfo.MediaInfo{
					Title: h.Title,
					Plot:  h.Description,
					Aired: nfo.Aired(h.Dates["broadcast_begin_date"].Time()),
					UniqueID: []nfo.ID{
						{
							ID:   strconv.Itoa(h.ID),
							Type: "FRANCETV:ID",
						},
						{
							ID:   h.SiID.String(),
							Type: "FRANCETV:SI_ID",
						},
					},
				}
				info.Actor = []nfo.Actor{}
				info.Tag = []string{}
				if len(h.Program.Label) > 0 {
					info.Showtitle = h.Program.Label
				}
				if len(h.Casting) > 0 && len(h.Characters) > 0 {
					actors := strings.Split(h.Casting, ",")
					characters := strings.Split(h.Characters, ",")

					for i := 0; i < len(actors); i++ {
						if i < len(characters) {
							info.Actor = append(info.Actor, nfo.Actor{Name: strings.TrimSpace(actors[i]), Role: strings.TrimSpace(characters[i]), Type: "Actor"})
						}
					}
				}

				if len(h.Presenter) > 0 {
					info.Actor = append(info.Actor, nfo.Actor{Name: h.Presenter, Type: "Presenter"})
				}

				if len(h.Director) > 0 {
					directors := strings.Split(h.Director, ",")
					for i := 0; i < len(directors); i++ {
						info.Actor = append(info.Actor, nfo.Actor{Name: strings.TrimSpace(directors[i]), Type: "Director"})
					}
				}

				if len(h.Producer) > 0 {
					producers := strings.Split(h.Producer, ",")
					for i := 0; i < len(producers); i++ {
						info.Actor = append(info.Actor, nfo.Actor{Name: strings.TrimSpace(producers[i]), Type: "Producer"})
					}
				}

				if len(h.Categories) > 0 {
					info.Genre = make([]string, len(h.Categories))
					for i := 0; i < len(h.Categories); i++ {
						info.Genre = append(info.Genre, h.Categories[i].Label)
					}
				}

				if len(h.Channels) > 0 {
					info.Tag = append(info.Tag, h.Channels[0].Label)
				}

				info.Season = h.SeasonNumber
				info.Episode = h.EpisodeNumber
				info.Thumb = make([]nfo.Thumb, 0)
				for k, format := range h.Image.Formats {
					url := ""
					maxW := 0
					for w, u := range format.Urls {
						width := 0
						_, err := fmt.Sscanf(w, "w:%d", &width)
						if err != nil {
							continue
						}
						if width > maxW {
							maxW = width
							url = u
						}
					}
					switch k {
					case "vignette_16x9":
						info.Thumb = append(info.Thumb, nfo.Thumb{Aspect: "thumb", URL: homeFranceTV + url})
					case "carre":
						info.Thumb = append(info.Thumb, nfo.Thumb{Aspect: "poster", URL: homeFranceTV + url})
					}
				}

				if media.ShowType == providers.Series {
					if info.Season == 0 {
						info.Season = info.Aired.Time().Year()
					}
				}
				mm <- media
			}
		}
	}()
	return mm
}
