package artetv

import (
	"encoding/json"
	"time"
)

// cSpell:disable

type APIResult struct {
	Datakey    Datakey `json:"datakey"`
	PageNumber int     `json:"pageNumber"`
	NextPage   string  `json:"nextPage"`
	Data       []Data  `json:"data"`
}
type Datakey struct {
	ID    string            `json:"id"`
	Param map[string]string `json:"param"`
}
type SearchAPI struct {
	ID                   string                 `json:"id"`
	Page                 string                 `json:"page"`
	Language             string                 `json:"language"`
	Support              string                 `json:"support"`
	Level                int                    `json:"level"`
	Parent               interface{}            `json:"parent"`
	AlternativeLanguages []AlternativeLanguages `json:"alternativeLanguages"`
	URL                  string                 `json:"url"`
	Deeplink             string                 `json:"deeplink"`
	Title                string                 `json:"title"`
	Description          string                 `json:"description"`
	Slug                 string                 `json:"slug"`
	Zones                []Zones                `json:"zones"`
}

type AlternativeLanguages struct {
	Code  string `json:"code"`
	Label string `json:"label"`
	Page  string `json:"page"`
	URL   string `json:"url"`
	Title string `json:"title"`
}
type Code struct {
	Name string      `json:"name"`
	ID   interface{} `json:"id"`
}
type DisplayOptions struct {
	ZoneLayout   string      `json:"zoneLayout"`
	ItemTemplate string      `json:"itemTemplate"`
	Theme        interface{} `json:"theme"`
}
type Kind struct {
	Code         string `json:"code"`
	Label        string `json:"label"`
	IsCollection bool   `json:"isCollection"`
}
type Resolutions struct {
	URL string `json:"url"`
	W   int    `json:"w"`
	H   int    `json:"h"`
}

type Image struct {
	Caption     string        `json:"caption"`
	Resolutions []Resolutions `json:"resolutions"`
	BlurURL     string        `json:"blurUrl"`
}

type Images map[string]Image

type Stickers struct {
	Code  string `json:"code"`
	Label string `json:"label"`
}

type Data struct {
	ID               string           `json:"id"`
	Type             string           `json:"type"`
	Kind             Kind             `json:"kind"`
	ProgramID        string           `json:"programId"`
	URL              string           `json:"url"`
	Deeplink         string           `json:"deeplink"`
	Title            string           `json:"title"`
	Subtitle         string           `json:"subtitle"`
	ShortDescription string           `json:"shortDescription"`
	Images           map[string]Image `json:"images"`
	Stickers         []Stickers       `json:"stickers"`
	Duration         interface{}      `json:"duration"`
	ChildrenCount    interface{}      `json:"childrenCount"`
	Geoblocking      interface{}      `json:"geoblocking"`
	Availability     interface{}      `json:"availability"`
	AgeRating        int              `json:"ageRating"`
}
type Zones struct {
	ID             string         `json:"id"`
	Code           Code           `json:"code"`
	Title          interface{}    `json:"title"`
	DisplayOptions DisplayOptions `json:"displayOptions"`
	Link           interface{}    `json:"link"`
	PageNumber     int            `json:"pageNumber"`
	NextPage       interface{}    `json:"nextPage"`
	Data           []Data         `json:"data"`
}

// tsGuide read broadcast time
var utcTZ, _ = time.LoadLocation("UTC")

type tsGuide time.Time

const tsGuidefmt = "2006-01-02T15:04:05Z"

func (t *tsGuide) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	d, err := time.ParseInLocation(tsGuidefmt, s, utcTZ)
	if err != nil {
		return err
	}
	*t = tsGuide(d)
	return nil
}

func (t tsGuide) MarshalJSON() ([]byte, error) {
	u := time.Time(t).Unix()
	return json.Marshal(u)
}

func (t tsGuide) Time() time.Time {
	return time.Time(t)
}

type tsAvailability time.Time

const tsAvailabilityfmt = "02/01/2006 15:04:05 -0700"

func (t *tsAvailability) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	d, err := time.ParseInLocation(tsAvailabilityfmt, s, utcTZ)
	if err != nil {
		return err
	}
	*t = tsAvailability(d)
	return nil
}

func (t tsAvailability) MarshalJSON() ([]byte, error) {
	u := time.Time(t).Unix()
	return json.Marshal(u)
}

func (t tsAvailability) Time() time.Time {
	return time.Time(t)
}

type playerAPI struct {
	VideoJSONPlayer VideoJSONPlayer `json:"videoJsonPlayer"`
}
type VTU struct {
	IUR string `json:"IUR"`
}

type StreamInfo struct {
	ID                  string `json:"id"`
	Quality             string `json:"quality"`
	Width               int    `json:"width"`
	Height              int    `json:"height"`
	MediaType           string `json:"mediaType"`
	MimeType            string `json:"mimeType"`
	Bitrate             int    `json:"bitrate"`
	URL                 string `json:"url"`
	VersionProg         int    `json:"versionProg"`
	VersionCode         string `json:"versionCode"`
	VersionLibelle      string `json:"versionLibelle"`
	VersionShortLibelle string `json:"versionShortLibelle"`
}

type VSR map[string]StreamInfo

type VideoJSONPlayer struct {
	VID                  string         `json:"VID"`
	VPI                  string         `json:"VPI"`
	VideoDurationSeconds int            `json:"videoDurationSeconds"`
	VideoIsoLang         string         `json:"videoIsoLang"`
	VTY                  string         `json:"VTY"`
	VTX                  string         `json:"VTX"`
	VTI                  string         `json:"VTI"`
	VDU                  int            `json:"VDU"`
	TcStartFrom          int            `json:"tc_start_from"`
	Autostart            bool           `json:"autostart"`
	LiveStreamRights     bool           `json:"liveStreamRights"`
	VGB                  string         `json:"VGB"`
	VRA                  tsAvailability `json:"VRA"`
	VRU                  tsAvailability `json:"VRU"`
	VAD                  bool           `json:"VAD"`
	VideoWarning         bool           `json:"videoWarning"`
	VTU                  VTU            `json:"VTU"`
	VTR                  string         `json:"VTR"`
	VUP                  string         `json:"VUP"`
	V7T                  string         `json:"V7T"`
	VDE                  string         `json:"VDE"`
	Postroll             string         `json:"postroll"`
	VSR                  VSR            `json:"VSR"`
	// Tracking             Tracking       `json:"tracking"`
	Platform       string `json:"platform"`
	VideoPlayerURL string `json:"videoPlayerUrl"`
	CaseProgram    string `json:"caseProgram"`
	// Categories           []Categories   `json:"categories"`
	// Category             Category       `json:"category"`
	// Subcategory          Subcategory    `json:"subcategory"`
	Language         string        `json:"language"`
	ProgramID        string        `json:"programId"`
	Genre            string        `json:"genre"`
	MainPlatformCode string        `json:"mainPlatformCode"`
	VSO              string        `json:"VSO"`
	Kind             string        `json:"kind"`
	KindLabel        string        `json:"kindLabel"`
	Collections      []interface{} `json:"collections"`
	ArteClub         bool          `json:"arteClub"`
	ProgramType      string        `json:"programType"`
	ParentProgramID  string        `json:"parentProgramId"`
	Markings         []interface{} `json:"markings"`
	// Adtech           Adtech        `json:"adtech"`
	// EStat            EStat         `json:"eStat"`
	// Smart            Smart         `json:"smart"`
	Illico        bool `json:"illico"`
	EnablePreroll bool `json:"enablePreroll"`
}
