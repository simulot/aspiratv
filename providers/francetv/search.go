package francetv

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/simulot/aspiratv/matcher"
	"github.com/simulot/aspiratv/media"
	"github.com/simulot/aspiratv/metadata/nfo"
	"github.com/simulot/aspiratv/net/myhttp/httptest"
	"github.com/simulot/aspiratv/providers/francetv/query"
)

const homeFranceTV = "https://www.france.tv"

type RequestPayLoad struct {
	Term    string   `json:"term"`
	Signal  Signal   `json:"signal"`
	Options *Options `json:"options,omitempty"`
}
type Signal struct {
}
type Options struct {
	ContentsLimit   int    `json:"contentsLimit,omitempty"`
	TaxonomiesLimit int    `json:"taxonomiesLimit,omitempty"`
	Types           string `json:"types"`
}

func (p *FranceTV) search(ctx context.Context, mr *matcher.MatchRequest) chan *media.Media {
	mm := make(chan *media.Media)

	go func() {
		var err error
		p.config.Log.Trace().Printf("[%s] Search for %q", p.Name(), mr.Show)

		defer func() {
			p.config.Log.Trace().Printf("[%s] Search for %q is done", p.Name(), mr.Show)
			if err != nil {
				p.config.Log.Error().Printf("[%s] Can't search: %w", p.Name(), err)

			}
			close(mm)
		}()
		// ctx, done := context.WithTimeout(ctx, p.deadline)
		// defer done()

		rq := RequestPayLoad{
			Term: mr.Show,
			Options: &Options{
				ContentsLimit: 20,
				// TaxonomiesLimit: 20,
				Types: "content",
			},
		}

		var resp []byte
		resp, err = json.Marshal(rq)
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
			p.config.Log.Debug().Printf(string(resp))
		}

		var r io.ReadCloser
		r, err = p.getter.DoWithContext(ctx, "POST", "https://www.france.tv/recherche/lancer/", h, bytes.NewBuffer(resp))
		if err != nil {
			p.config.Log.Error().Printf("[%s] Can't call search API: %s", p.Name(), err)
			return
		}
		if p.config.Log.IsDebug() {
			r = httptest.DumpReaderToFile(p.config.Log, r, "francetv-recherche-")
		}

		resp, err = ioutil.ReadAll(r)
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
			reEncode, err := json.MarshalIndent(results, "", "  ")
			if err != nil {
				p.config.Log.Error().Printf("Can't encode json response: %s", err)
			}
			p.config.Log.Debug().Printf("[%s] Decoded result\n%s", p.Name(), string(reEncode))
		}

		// Search for series first
		series := map[int]query.Program{}
		for _, cat := range []string{"content", "taxonomy"} {
			for _, hit := range results[cat].Hits {
				if strings.Contains(strings.ToLower(hit.Program.Label), mr.Show) {
					series[hit.Program.ID] = hit.Program
				}
			}
		}

		if len(series) > 0 {
			for _, prog := range series {
				p.visitPageSerie(ctx, mr, mm, prog.URLComplete)
			}
			return
		}

		// Other videos

		for _, result := range results {
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

				media := &media.Media{
					ID:    h.SiID.String(),
					Match: mr,
				}
				var info *nfo.MediaInfo

				if h.SeasonNumber != 0 || h.Class == "program" || h.Program.Class == "program" {
					meta := nfo.EpisodeDetails{}
					info = &meta.MediaInfo
					info.MediaType = nfo.TypeSeries
					media.SetMetaData(&meta)

				} else {
					meta := nfo.Movie{}
					info = &meta.MediaInfo
					media.SetMetaData(&meta)
					info.MediaType = nfo.TypeMovie
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

				if info.MediaType == nfo.TypeSeries {
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

var reID = regexp.MustCompile(`\/(\d+)-[^\/]+\.html$`)
var reAnalyseTitle = regexp.MustCompile(`^\s?S(\d+)?\s+E(\d+)\s+-\s+(.*)$`)
var reAired = regexp.MustCompile(`(\d{2})\/(\d{2})`)

func (p *FranceTV) visitPageSerie(ctx context.Context, mr *matcher.MatchRequest, mm chan *media.Media, url string) error {
	// https://www.france.tv/series-et-fictions/series-policieres-thrillers/district-31
	// https://www.france.tv/recherche/lancer/query=district+31\u0026hitsPerPage=20\u0026page=0\u0026filters=(class%3Aprogram%20OR%20class%3Aevent)%20AND%20(counters.web.integral_counter%20%3E%200%20OR%20counters.web.extract_counter%20%3E%200)%20AND%20NOT%20type%3Asaison%20AND%20NOT%20type%3Acomposite\u0026restrictSearchableAttributes=%5B%22label%22%2C%22title%22%2C%22description%22%2C%22seo%22%5D

	parser := p.htmlParserFactory.New()
	page := 0
	hits := 0
	lastPageWithHits := 0

	parser.OnHTML("a.c-card-video", func(e *colly.HTMLElement) {
		if strings.Contains(e.Attr("class"), "unavailable") {
			return
		}

		showTitle := e.ChildText("span.c-card-video__textarea-title")
		if !strings.Contains(strings.ToLower(showTitle), mr.Show) {
			return
		}

		u := e.Attr("href")
		id := ""

		match := reID.FindStringSubmatch(u)
		if len(match) == 2 {
			id = match[1]
		}

		info := nfo.MediaInfo{
			UniqueID: []nfo.ID{
				{
					ID:   id,
					Type: "francetv",
				},
			},
			PageURL:   "https://www.france.tv/" + u,
			Showtitle: showTitle,
		}

		subtitle := e.ChildText("span.c-card-video__textarea-subtitle")
		if match = reAired.FindStringSubmatch(e.Text); len(match) == 4 {
			info.Season, _ = strconv.Atoi(match[1])
			info.Episode, _ = strconv.Atoi(match[2])
			info.Title = strings.TrimSpace(match[3])
			info.MediaType = nfo.TypeSeries

		} else {
			info.Title = subtitle
			info.MediaType = nfo.TypeShow
		}

		e.ForEach("span.c-metadata", func(n int, e *colly.HTMLElement) {
			if match = reAired.FindStringSubmatch(e.Text); len(match) == 3 {
				day, _ := strconv.Atoi(match[1])
				month, _ := strconv.Atoi(match[2])
				year := time.Now().Year()
				d := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
				if d.After(time.Now()) {
					d = time.Date(year-1, time.Month(month), day, 0, 0, 0, 0, time.Local)
				}
				info.Aired = nfo.Aired(d)
			}
		})

		switch e.ChildText("span.c-label") {
		case "extrait":
			info.IsBonus = true
		}

		if !mr.KeepBonus && info.IsBonus {
			return
		}
		p.config.Log.Trace().Printf("[%s] Found %q", p.Name(), info.Title)

		media := &media.Media{
			ID:    id,
			Match: mr,
			Metadata: &nfo.EpisodeDetails{
				MediaInfo: info,
			},
		}
		hits++
		lastPageWithHits = page
		mm <- media
	})

	url = "https://www.france.tv/" + url + "/toutes-les-videos/"

	for {
		u := url + "?page=" + strconv.Itoa(page)

		p.config.Log.Trace().Printf("[%s] Visiting page %q", p.Name(), u)
		err := parser.Visit(u)
		if err != nil {
			return err
		}
		if hits == 0 && page-lastPageWithHits > 1 {
			break
		}
		page++
		hits = 0
	}
	return nil
}
