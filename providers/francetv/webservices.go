package francetv

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type seasonWrapper struct {
	Season
	Partial bool `json:"-"`
}

type intOrString string

type QueryResults struct {
	Results map[string]Result
}

type Format struct {
	OriginalPath string            `json:"original_path"`
	OriginalName string            `json:"original_name"`
	Urls         map[string]string `json:"urls"`
}
type Image struct {
	ID      int               `json:"id"`
	Title   string            `json:"title"`
	Credit  string            `json:"credit"`
	Formats map[string]Format `json:"formats"`
}

type Season struct {
	ID           int    `json:"id"`
	Class        string `json:"class"`
	Type         string `json:"type"`
	Label        string `json:"label"`
	URL          string `json:"url"`
	URLComplete  string `json:"url_complete"`
	Season       int    `json:"season"`
	EpisodeCount int    `json:"episode_count"`
	Logo         Logo   `json:"logo"`
}
type Categories struct {
	ID           int         `json:"id"`
	Class        string      `json:"class"`
	Type         string      `json:"type"`
	Label        string      `json:"label"`
	URL          string      `json:"url"`
	URLComplete  string      `json:"url_complete"`
	Season       interface{} `json:"season"`
	EpisodeCount interface{} `json:"episode_count"`
}
type Channels struct {
	ID           int    `json:"id"`
	Class        string `json:"class"`
	Type         string `json:"type"`
	Label        string `json:"label"`
	URL          string `json:"url"`
	URLComplete  string `json:"url_complete"`
	Season       int    `json:"season"`
	EpisodeCount int    `json:"episode_count"`
}

type Program struct {
	ID           int    `json:"id"`
	Class        string `json:"class"`
	Type         string `json:"type"`
	Label        string `json:"label"`
	URL          string `json:"url"`
	URLComplete  string `json:"url_complete"`
	Season       int    `json:"season"`
	EpisodeCount int    `json:"episode_count"`
	Logo         Logo   `json:"logo"`
}

type Logo struct {
	OriginalName string `json:"original_name"`
	OriginalPath string `json:"original_path"`
	ID           int    `json:"id"`
}
type Hits struct {
	ID                      int                      `json:"id,omitempty"`
	Class                   string                   `json:"class,omitempty"`
	Type                    string                   `json:"type,omitempty"`
	Label                   string                   `json:"label,omitempty"`
	Title                   string                   `json:"title,omitempty"`
	HeadlineTitle           string                   `json:"headline_title,omitempty"`
	Description             string                   `json:"description,omitempty"`
	Text                    string                   `json:"text,omitempty"`
	URLPage                 string                   `json:"url_page,omitempty"`
	Path                    string                   `json:"path,omitempty"`
	Duration                Duration                 `json:"duration,omitempty"`
	SeasonNumber            int                      `json:"season_number,omitempty"`
	EpisodeNumber           int                      `json:"episode_number,omitempty"`
	IsAudioDescripted       bool                     `json:"is_audio_descripted,omitempty"`
	IsPreviouslyBroadcasted bool                     `json:"is_previously_broadcasted,omitempty"`
	IsMultiLingual          bool                     `json:"is_multi_lingual,omitempty"`
	IsSubtitled             bool                     `json:"is_subtitled,omitempty"`
	IsPreview               bool                     `json:"is_preview,omitempty"`
	IsSponsored             bool                     `json:"is_sponsored,omitempty"`
	Director                string                   `json:"director,omitempty"`
	Producer                string                   `json:"producer,omitempty"`
	Presenter               string                   `json:"presenter,omitempty"`
	Casting                 string                   `json:"casting,omitempty"`
	Characters              string                   `json:"characters,omitempty"`
	ProductionYear          int                      `json:"production_year,omitempty"`
	Dates                   map[string]UnixTimeStamp `json:"dates,omitempty"`
	URL                     string                   `json:"url,omitempty"`
	URLComplete             string                   `json:"url_complete,omitempty"`
	Synopsis                string                   `json:"synopsis,omitempty"`
	Image                   Image                    `json:"image,omitempty"`
	Categories              []Categories             `json:"categories,omitempty"`
	Channels                []Channels               `json:"channels,omitempty"`
	Program                 Program                  `json:"program,omitempty"`
	Season                  seasonWrapper            `json:"season,omitempty"`
	RatingCsaCode           string                   `json:"rating_csa_code,omitempty"`
	SiID                    intOrString              `json:"si_id,omitempty"`
	ObjectID                string                   `json:"objectID,omitempty"`
	VideoCount              int                      `json:"video_count,omitempty"`
	Channel                 string                   `json:"channel,omitempty"`
	// FreeID        int        `json:"free_id"`
	// Ranges                  Ranges          `json:"ranges"`
	// OrangeID      string     `json:"orange_id"`
	// HighlightResult         HighlightResult `json:"_highlightResult"`
	// Parent *Hits `json:"parent"`
}
type Result struct {
	Hits             []Hits `json:"hits"`
	NbHits           int    `json:"nbHits"`
	Page             int    `json:"page"`
	NbPages          int    `json:"nbPages"`
	HitsPerPage      int    `json:"hitsPerPage"`
	ProcessingTimeMS int    `json:"processingTimeMS"`
	ExhaustiveNbHits bool   `json:"exhaustiveNbHits"`
	Query            string `json:"query"`
	Params           string `json:"params"`
	Index            string `json:"index"`
}

type UnixTimeStamp time.Time

func (v *UnixTimeStamp) UnmarshalJSON(b []byte) error {
	s := string(b)
	if s == "null" {
		*v = UnixTimeStamp(time.Time{})
		return nil
	}
	i, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return fmt.Errorf("Can't convert unix timestamp: %w", err)
	}
	*v = UnixTimeStamp(time.Unix(i, 0))
	return nil
}

func (v UnixTimeStamp) Time() time.Time {
	return time.Time(v)
}

type Duration time.Duration

func (v *Duration) UnmarshalJSON(b []byte) error {
	var d int64
	err := json.Unmarshal(b, &d)
	if err != nil {
		return fmt.Errorf("Can't convert duration: %w", err)
	}
	*v = Duration(time.Duration(d) * time.Second)
	return nil
}

func (v Duration) Duration() time.Duration { return time.Duration(v) }

func (v *seasonWrapper) UnmarshalJSON(b []byte) error {
	if s, err := strconv.Atoi(string(b)); err == nil {
		v.Partial = true
		v.Season.Season = s
		return nil
	}
	return json.Unmarshal(b, &v.Season)
}

func (v *intOrString) UnmarshalJSON(b []byte) error {
	if _, err := strconv.Atoi(string(b)); err == nil {
		*v = intOrString(string(b))
		return nil
	}
	if len(b) > 2 {
		*v = intOrString(string(b[1 : len(b)-1]))
	}
	return nil

}

func (v intOrString) String() string {
	return string(v)
}

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
