package query

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
	Results []Results `json:"results"`
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
	ID                      int                      `json:"id"`
	Class                   string                   `json:"class"`
	Type                    string                   `json:"type"`
	Label                   string                   `json:"label"`
	Title                   string                   `json:"title"`
	HeadlineTitle           string                   `json:"headline_title"`
	Description             string                   `json:"description"`
	Text                    string                   `json:"text"`
	URLPage                 string                   `json:"url_page"`
	Path                    string                   `json:"path"`
	Duration                Duration                 `json:"duration"`
	SeasonNumber            int                      `json:"season_number"`
	EpisodeNumber           int                      `json:"episode_number"`
	IsAudioDescripted       bool                     `json:"is_audio_descripted"`
	IsPreviouslyBroadcasted bool                     `json:"is_previously_broadcasted"`
	IsMultiLingual          bool                     `json:"is_multi_lingual"`
	IsSubtitled             bool                     `json:"is_subtitled"`
	IsPreview               bool                     `json:"is_preview"`
	IsSponsored             bool                     `json:"is_sponsored"`
	Director                string                   `json:"director"`
	Producer                string                   `json:"producer"`
	Presenter               string                   `json:"presenter"`
	Casting                 string                   `json:"casting"`
	Characters              string                   `json:"characters"`
	ProductionYear          int                      `json:"production_year"`
	Dates                   map[string]UnixTimeStamp `json:"dates"`
	// Ranges                  Ranges          `json:"ranges"`
	Image         Image         `json:"image,omitempty"`
	Categories    []Categories  `json:"categories"`
	Channels      []Channels    `json:"channels"`
	Program       Program       `json:"program"`
	Season        seasonWrapper `json:"season"`
	RatingCsaCode string        `json:"rating_csa_code"`
	SiID          intOrString   `json:"si_id"`
	// FreeID        int        `json:"free_id"`
	// OrangeID      string     `json:"orange_id"`
	ObjectID string `json:"objectID"`
	// HighlightResult         HighlightResult `json:"_highlightResult"`
	// Parent *Hits `json:"parent"`
}
type Results struct {
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
