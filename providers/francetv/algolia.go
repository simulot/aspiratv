package francetv

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

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

const homeFranceTV = "https://www.france.tv/"

func (p *FranceTV) getAlgoliaConfig(ctx context.Context) error {

	r, err := p.getter.Get(ctx, homeFranceTV)
	if err != nil {
		return fmt.Errorf("Can't get FranceTV home page :%w", err)
	}

	defer r.Close()

	if p.debug {
		r = httptest.DumpReaderToFile(r, "franctv-home-")
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

func (p *FranceTV) queryAlgolia(ctx context.Context, m *providers.MatchRequest) chan *providers.Show {
	shows := make(chan *providers.Show)

	go func() {
		defer close(shows)
		// ctx, done := context.WithTimeout(ctx, p.deadline)
		// defer done()

		const algoliaURL = "https://vwdlashufe-dsn.algolia.net/1/indexes/*/queries"

		v := url.Values{}
		v.Set("x-algolia-agent", "Algolia for vanilla JavaScript (lite) 3.27.0;instantsearch.js 2.10.2;JS Helper 2.26.0")
		v.Set("x-algolia-application-id", p.algolia.AlgoliaAppID)
		v.Set("x-algolia-api-key", p.algolia.AlgoliaAPIKey)

		u := algoliaURL + "?" + v.Encode()

		if p.debug {
			log.Printf("[%s] Search url %q", p.Name(), u)
		}
		page := 0
		ts := time.Now().Unix()
		req := AlgoliaParam{
			"query":        m.Show,
			"hitsPerPage":  "20",
			"filters":      fmt.Sprintf("class:video AND ranges.replay.web.begin_date < %d AND ranges.replay.web.end_date > %d", ts, ts),
			"facetFilters": `[["class:video"]]`,
			"facets":       "[]",
			"tagFilters":   "",
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
			if p.debug {
				log.Printf("[%s] Request body", p.Name())
				for k, s := range h {
					log.Printf("%q %s", k, strings.Join(s, ","))
				}
				log.Println(b.String())
			}

			r, err := p.getter.DoWithContext(ctx, "POST", u, h, b)
			if err != nil {
				log.Printf("[%s] Can't call algolia API: %s", p.Name(), err)
				return
			}
			if p.debug {
				r = httptest.DumpReaderToFile(r, "francetv-algolia-")
			}

			resp, err := ioutil.ReadAll(r)
			if err != nil {
				log.Printf("[%s] Can't get API result: %s", p.Name(), err)
				return
			}
			results := query.QueryResults{}
			err = json.Unmarshal(resp, &results)
			if err != nil {
				log.Printf("[%s] Can't decode API result: %s", p.Name(), err)
				return
			}
			r.Close()
			for resNum := range results.Results {
				for _, h := range results.Results[resNum].Hits {
					if h.Type != "integrale" {
						continue
					}
					st := strings.ToLower(h.Program.Label)
					_ = st
					if len(h.Program.Label) > 0 && !strings.Contains(strings.ToLower(h.Program.Label), m.Show) {
						continue
					}

					if len(h.Program.Label) == 0 && !strings.Contains(strings.ToLower(h.Title), m.Show) {
						continue
					}

					// ID:           strconv.Itoa(h.ID),
					s := providers.Show{
						Show:         h.Program.Label,
						Title:        h.Title,
						ID:           h.SiID,
						Pitch:        h.Description,
						AirDate:      h.Dates["broadcast_begin_date"].Time(),
						Provider:     p.Name(),
						ThumbnailURL: homeFranceTV + h.Image.Formats["carre"].Urls["w:400"],
						Destination:  m.Destination,
					}
					if h.EpisodeNumber > 0 {
						s.Episode = strconv.Itoa(h.EpisodeNumber)
					}
					if h.SeasonNumber > 0 {
						s.Season = strconv.Itoa(h.SeasonNumber)
					}
					if len(h.Channels) > 0 {
						s.Channel = h.Channels[0].Label
					}
					s.ShowURL = homeFranceTV + h.Path + "/" + s.ID + "-" + h.URLPage + ".html"
					shows <- &s
				}
			}
			page++
			if page >= results.Results[0].NbPages {
				break
			}
		}
	}()
	return shows
}

func encodeRequest(b *bytes.Buffer, w *algoliaRequestWrapper) {
	b.WriteByte('{')
	b.WriteString(`"requests":[{`)
	b.WriteString(`"indexName":"yatta_prod_contents","params":`)
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
