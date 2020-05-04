package francetv

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/simulot/aspiratv/metadata/nfo"
	"github.com/simulot/aspiratv/net/myhttp/httptest"
	"github.com/simulot/aspiratv/providers"
	"github.com/simulot/aspiratv/providers/francetv/query"
)

// AlgoliaConfig to be extracted from home page
type AlgoliaConfig struct {
	AlgoliaAPIContentMaxPage            int    `json:"algolia_api_content_max_page"`
	AlgoliaAPIContentPaginitationLimit  int    `json:"algolia_api_content_paginitation_limit"`
	AlgoliaAPIIndexContent              string `json:"algolia_api_index_content"`
	AlgoliaAPIIndexTaxonomy             string `json:"algolia_api_index_taxonomy"`
	AlgoliaAPIKey                       string `json:"algolia_api_key"`
	AlgoliaAPITaxonomyPaginitationLimit int    `json:"algolia_api_taxonomy_paginitation_limit"`
	AlgoliaAppID                        string `json:"algolia_app_id"`
	BookmarkGetURL                      string `json:"bookmark_get_url"`
	BookmarkPostURL                     string `json:"bookmark_post_url"`
	Environment                         string `json:"environment"`
	GinLibURL                           string `json:"gin_lib_url"`
	GinPersonalDataLink                 string `json:"gin_personal_data_link"`
	RecoSystemAuthorization             string `json:"reco_system_authorization"`
	RecoSystemHost                      string `json:"reco_system_host"`
	TagManagementSystemsURL             string `json:"tag_management_systems_url"`
	WatchingGetURL                      string `json:"watching_get_url"`
	WatchingHitTimer                    int    `json:"watching_hit_timer"`
	WatchingMinTime                     int    `json:"watching_min_time"`
	WatchingPostURL                     string `json:"watching_post_url"`
}

var algoliaRegexp = regexp.MustCompile(`getAppConfig\(\)\s*\{\s*return\s+(\{[^;]+\})\s*;`)

const homeFranceTV = "https://www.france.tv"

func (p *FranceTV) getAlgoliaConfig(ctx context.Context) error {

	r, err := p.getter.Get(ctx, homeFranceTV)
	if err != nil {
		return fmt.Errorf("Can't get FranceTV home page :%w", err)
	}

	defer r.Close()

	if p.config.Log.IsDebug() {
		r = httptest.DumpReaderToFile(p.config.Log, r, "francetv-home-")
	}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Errorf("Can't get FranceTV home page :%w", err)
	}

	m := algoliaRegexp.FindAllSubmatchIndex(b, -1)
	if len(m) < 1 || len(m[0]) < 4 {
		return errors.New("Can't find Algolia configuration")
	}

	conf := &AlgoliaConfig{}
	err = json.Unmarshal(b[m[0][2]:m[0][3]], conf)
	if err != nil {
		return fmt.Errorf("Can't decode Algolia configuration: %w", err)
	}
	p.algolia = conf
	return nil
}

type algoliaRequestWrapper struct {
	Requests []Requests `json:"requests"`
}
type Requests struct {
	IndexName string       `json:"indexName"`
	Params    AlgoliaParam `json:"params"`
}

const algoliaURL = "https://vwdlashufe-dsn.algolia.net/1/indexes/*/queries"

func (p *FranceTV) queryAlgolia(ctx context.Context, mr *providers.MatchRequest) chan *providers.Media {
	mm := make(chan *providers.Media)

	go func() {
		defer close(mm)
		// ctx, done := context.WithTimeout(ctx, p.deadline)
		// defer done()

		v := url.Values{}
		v.Set("x-algolia-agent", "Algolia for vanilla JavaScript (lite) 3.27.0;instantsearch.js 2.10.2;JS Helper 2.26.0")
		v.Set("x-algolia-application-id", p.algolia.AlgoliaAppID)
		v.Set("x-algolia-api-key", p.algolia.AlgoliaAPIKey)

		u := algoliaURL + "?" + v.Encode()

		p.config.Log.Debug().Printf("[%s] Search url %q", p.Name(), u)
		page := 0
		ts := time.Now().Unix()
		req := AlgoliaParam{
			"query":        mr.Show,
			"hitsPerPage":  "20",
			"filters":      fmt.Sprintf("class:video AND ranges.replay.web.begin_date < %d AND ranges.replay.web.end_date > %d", ts, ts),
			"facetFilters": `[["class:video"]]`,
			"facets":       "[]",
			"tagFilters":   "",
		}
		if mr.MaxAgedDays > 0 {
			fromTS := time.Now().AddDate(0, 0, -mr.MaxAgedDays-1).Unix()
			req["filters"] += fmt.Sprintf(" AND dates.broadcast_begin_date > %d", fromTS)
		}

		for {
			req["page"] = strconv.Itoa(page)
			w := algoliaRequestWrapper{
				Requests: []Requests{
					{
						IndexName: "yatta_prod_contents",
						Params:    req,
					},
				},
			}
			_ = w
			b := bytes.NewBuffer([]byte{})
			encodeRequest(b, &w) // Special encoding... WTF

			h := make(http.Header)
			h.Add("Accept", "application/json")
			h.Add("Accept-Language", "fr-FR,fr;q=0.5")
			h.Add("Accept-Encoding", "gzip")
			h.Add("Referer", "https://www.france.tv")
			h.Add("content-type", "https://www.france.tv")
			h.Add("Origin", "https://www.france.tv")
			h.Add("TE", "Trailers")
			if p.config.Log.IsDebug() {
				p.config.Log.Debug().Printf("[%s] Request headers", p.Name())
				for k, s := range h {
					p.config.Log.Debug().Printf("[%s] %q %s", p.Name(), k, strings.Join(s, ","))
				}
				p.config.Log.Debug().Printf(b.String())
			}

			r, err := p.getter.DoWithContext(ctx, "POST", u, h, b)
			if err != nil {
				p.config.Log.Error().Printf("[%s] Can't call algolia API: %s", p.Name(), err)
				return
			}
			if p.config.Log.IsDebug() {
				r = httptest.DumpReaderToFile(p.config.Log, r, "francetv-algolia-")
			}

			resp, err := ioutil.ReadAll(r)
			if err != nil {
				p.config.Log.Error().Printf("[%s] Can't get API result: %s", p.Name(), err)
				return
			}
			results := query.QueryResults{}
			err = json.Unmarshal(resp, &results)
			if err != nil {
				p.config.Log.Error().Printf("[%s] Can't decode API result: %s", p.Name(), err)
				return
			}
			r.Close()
			for resNum := range results.Results {
				for _, h := range results.Results[resNum].Hits {
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

					if len(h.Program.Label) > 0 {
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
						info.SeasonInfo, info.TVShow = p.getProgram(ctx, info.Showtitle, h.Season.ID, h.Program.ID)
						if !info.IsSpecial {
							if info.Season == 0 {
								info.Season = info.Aired.Time().Year()
							}
						}
					}
					mm <- media
				}
			}
			page++
			if page >= results.Results[0].NbPages {
				break
			}
		}
	}()
	return mm
}

func (p *FranceTV) getProgram(ctx context.Context, program string, seasonID, programID int) (*nfo.Season, *nfo.TVShow) {

	season, ok1 := p.seasons.Load(seasonID)
	show, ok2 := p.shows.Load(programID)

	if ok1 && ok2 {
		return season.(*nfo.Season), show.(*nfo.TVShow)
	}

	// program = strings.ToLower(program)
	v := url.Values{}
	v.Set("x-algolia-agent", "Algolia for vanilla JavaScript (lite) 3.27.0;instantsearch.js 2.10.2;JS Helper 2.26.0")
	v.Set("x-algolia-application-id", p.algolia.AlgoliaAppID)
	v.Set("x-algolia-api-key", p.algolia.AlgoliaAPIKey)

	u := algoliaURL + "?" + v.Encode()
	p.config.Log.Debug().Printf("[%s] Search url %q", p.Name(), u)
	page := 0
	req := AlgoliaParam{
		"query":       program,
		"hitsPerPage": "20",
		// "filters":      fmt.Sprintf("class:program AND (counters.web.integral_counter > 0 OR counters.web.extract_counter > 0)"),
		"filters":      fmt.Sprintf("class:program"),
		"facetFilters": `[["class:program"]]`,
		"facets":       "[]",
		"tagFilters":   "",
	}
	for {
		req["page"] = strconv.Itoa(page)
		w := algoliaRequestWrapper{
			Requests: []Requests{
				{
					IndexName: "yatta_prod_taxonomies",
					Params:    req,
				},
			},
		}
		_ = w
		b := bytes.NewBuffer([]byte{})
		encodeRequest(b, &w) // Special encoding... WTF

		h := make(http.Header)
		h.Add("Accept", "application/json")
		h.Add("Accept-Language", "fr-FR,fr;q=0.5")
		h.Add("Accept-Encoding", "gzip")
		h.Add("Referer", "https://www.france.tv")
		h.Add("content-type", "https://www.france.tv")
		h.Add("Origin", "https://www.france.tv")
		h.Add("TE", "Trailers")
		if p.config.Log.IsDebug() {
			p.config.Log.Debug().Printf("[%s] Request headers", p.Name())
			for k, s := range h {
				p.config.Log.Debug().Printf("%q %s", k, strings.Join(s, ","))
			}
		}

		r, err := p.getter.DoWithContext(ctx, "POST", u, h, b)
		if err != nil {
			p.config.Log.Error().Printf("[%s] Can't call algolia API: %s", p.Name(), err)
			return nil, nil
		}
		if p.config.Log.IsDebug() {
			r = httptest.DumpReaderToFile(p.config.Log, r, "francetv-algolia-pgm-")
		}

		resp, err := ioutil.ReadAll(r)
		if err != nil {
			p.config.Log.Error().Printf("[%s] Can't get API result: %s", p.Name(), err)
			return nil, nil
		}
		results := query.QueryResults{}
		err = json.Unmarshal(resp, &results)
		if err != nil {
			p.config.Log.Error().Printf("[%s] Can't decode API result: %s", p.Name(), err)
			return nil, nil
		}
		r.Close()

		for resNum := range results.Results {
			for _, h := range results.Results[resNum].Hits {
				if h.Class != "program" {
					continue
				}

				thumbs := []nfo.Thumb{}
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
					case "logo":
						thumbs = append(thumbs, nfo.Thumb{Aspect: "clearlogo", URL: homeFranceTV + url})
					case "vignette_16x9":
						thumbs = append(thumbs, nfo.Thumb{Aspect: "fanart", URL: homeFranceTV + url})
					case "carre":
						thumbs = append(thumbs, nfo.Thumb{Aspect: "poster", URL: homeFranceTV + url})
					case "background_16x9":
						thumbs = append(thumbs, nfo.Thumb{Aspect: "backdrop", URL: homeFranceTV + url})
					}
				}

				switch h.Type {
				case "program":
					p.shows.Store(h.ID, &nfo.TVShow{
						Title: h.Label,
						Plot:  h.Description,
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
						Thumb: thumbs,
					})
				case "saison":
					p.seasons.Store(h.ID, &nfo.Season{
						Title: h.Label,
						Plot:  h.Description,
						Thumb: thumbs,
					})

				}
			}
		}
		page++
		if page >= results.Results[0].NbPages {
			break
		}
	}

	var (
		theSeason *nfo.Season
		theShow   *nfo.TVShow
	)

	if season, ok1 = p.seasons.Load(seasonID); ok1 {
		theSeason = season.(*nfo.Season)
	}

	if show, ok2 = p.shows.Load(programID); ok2 {
		theShow = show.(*nfo.TVShow)
	}

	return theSeason, theShow
}

func encodeRequest(b *bytes.Buffer, w *algoliaRequestWrapper) {
	b.WriteByte('{')
	b.WriteString(`"requests":[{`)
	b.WriteString(`"indexName":"`)
	b.WriteString(w.Requests[0].IndexName)
	b.WriteString(`","params":`)
	b.WriteByte('"')
	shouldWriteAmp := false
	for k, v := range w.Requests[0].Params {
		if shouldWriteAmp {
			b.WriteByte('&')
		}
		encodeStringRequest(b, k)
		b.WriteByte('=')
		encodeStringRequest(b, v)
		shouldWriteAmp = true
	}
	b.WriteByte('"')
	b.WriteString(`}]}`)
}

type AlgoliaParam map[string]string

func encodeStringRequest(b *bytes.Buffer, s string) {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' {
			b.WriteByte(c)
			continue
		}
		switch c {
		case '_', '.', '$', '+', ',', '/', ';', '=', '?', '@':
			b.WriteByte(c)
			continue
		}

		encodeHex(b, c)
	}
}

func encodeHex(b *bytes.Buffer, c byte) {
	b.WriteByte('%')
	b.WriteByte("0123456789ABCDEF"[c>>4])
	b.WriteByte("0123456789ABCDEF"[c&15])
}
