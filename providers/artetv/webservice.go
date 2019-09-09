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

// Day guide structure
// Commented out fields are kept for documentation
// type guide struct {
// AlternativeLanguages []struct {
// 	Code  string `json:"code"`
// 	Label string `json:"label"`
// 	Page  string `json:"page"`
// 	Title string `json:"title"`
// 	URL   string `json:"url"`
// } `json:"alternativeLanguages"`
// Description string      `json:"description"`
// ID          string      `json:"id"`
// Images      interface{} `json:"images"`
// Language    string      `json:"language"`
// Level       int         `json:"level"`
// Page        string      `json:"page"`
// Parent      interface{} `json:"parent"`
// Slug        string      `json:"slug"`
// Stats       struct {
// 	Xiti struct {
// 		Chapter1       interface{} `json:"chapter1"`
// 		Chapter2       interface{} `json:"chapter2"`
// 		Chapter3       interface{} `json:"chapter3"`
// 		EnvWork        string      `json:"env_work"`
// 		PageName       string      `json:"page_name"`
// 		S2             int         `json:"s2"`
// 		SearchKeywords interface{} `json:"search_keywords"`
// 		SiteID         string      `json:"siteId"`
// 		X1             string      `json:"x1"`
// 		X2             string      `json:"x2"`
// 	} `json:"xiti"`
// } `json:"stats"`
// Support string `json:"support"`
// Title   string `json:"title"`
// URL     string `json:"url"`
// 	Zones []zones `json:"zones"`
// }

// type zones struct {
// 	Code code `json:"code"`
// 	// ContextPage    string `json:"contextPage"`
// 	Data []data `json:"data"`
// 	// DisplayOptions struct {
// 	// 	Layout   string      `json:"layout"`
// 	// 	Template string      `json:"template"`
// 	// 	Theme    interface{} `json:"theme"`
// 	// } `json:"displayOptions"`
// 	// Link     interface{} `json:"link"`
// 	// NextPage interface{} `json:"nextPage"`
// 	// Title    string      `json:"title"`
// 	// Type     string      `json:"type"`
// }

// type code struct {
// 	ID   interface{} `json:"id"`   // 0 for highlights, 1  for listing
// 	Name string      `json:"name"` // highlights_TV_GUIDE  listing_TV_GUIDE
// }

// type data struct {
// 	// Availability    interface{}   `json:"availability"`
// 	BroadcastDates []tsGuide `json:"broadcastDates"`
// 	// Credits         []interface{} `json:"credits"`
// 	Duration        jsonparser.Seconds `json:"duration"`
// 	Description     string             `json:"description"`
// 	FullDescription string             `json:"fullDescription"`
// 	// Geoblocking     interface{}   `json:"geoblocking"`
// 	ID     string            `json:"id"`
// 	Images map[string]thumbs `json:"images"`
// 	Kind   struct {
// 		Code  string      `json:"code"`
// 		Label interface{} `json:"label"`
// 	} `json:"kind"`
// 	// LivestreamRights    interface{} `json:"livestreamRights"`
// 	// OfflineAvailability interface{} `json:"offlineAvailability"`
// 	// Partners            interface{} `json:"partners"`
// 	// Player              struct {
// 	// 	HTML        string      `json:"html"`
// 	// 	LinkAndroid interface{} `json:"linkAndroid"`
// 	// 	LinkIos     interface{} `json:"linkIos"`
// 	// } `json:"player"`
// 	ProgramID string `json:"programId"`
// 	// ShopURL          interface{}   `json:"shopUrl"`
// 	ShortDescription string `json:"shortDescription"`
// 	// Stickers         []interface{} `json:"stickers"`
// 	Subtitle string `json:"subtitle"`
// 	Title    string `json:"title"`
// 	Type     string `json:"type"`
// 	URL      string `json:"url"`
// }

// type thumbs struct {
// 	BlurURL     string       `json:"blurUrl"`
// 	Caption     string       `json:"caption"`
// 	Resolutions []resolution `json:"resolutions"`
// }

// type resolution struct {
// 	Height int    `json:"height"`
// 	URL    string `json:"url"`
// 	Width  int    `json:"width"`
// }

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

// type VSR struct {
// 	HTTPSHQ2 HTTPSHQ2 `json:"HTTPS_HQ_2"`
// 	HTTPSEQ2 HTTPSEQ2 `json:"HTTPS_EQ_2"`
// 	HTTPSMQ2 HTTPSMQ2 `json:"HTTPS_MQ_2"`
// 	HTTPSSQ2 HTTPSSQ2 `json:"HTTPS_SQ_2"`
// 	HLSXQ2   HLSXQ2   `json:"HLS_XQ_2"`
// 	HTTPSMQ1 HTTPSMQ1 `json:"HTTPS_MQ_1"`
// 	HTTPSEQ1 HTTPSEQ1 `json:"HTTPS_EQ_1"`
// 	HTTPSHQ1 HTTPSHQ1 `json:"HTTPS_HQ_1"`
// 	HTTPSSQ1 HTTPSSQ1 `json:"HTTPS_SQ_1"`
// 	HLSXQ1   HLSXQ1   `json:"HLS_XQ_1"`
// }
type VSR map[string]StreamInfo

/*
func (p *VSR) UnmarshalJSON(b []byte) error {
	s := map[string]Stream{}

	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*p = s
	return nil
}
*/

// type Tablet struct {
// 	WEB   string `json:"WEB"`
// 	EMBED string `json:"EMBED"`
// }
// type Desktop struct {
// 	WEB   string `json:"WEB"`
// 	EMBED string `json:"EMBED"`
// }
// type Mobile struct {
// 	WEB   string `json:"WEB"`
// 	EMBED string `json:"EMBED"`
// }
// type Tracking struct {
// 	Tablet  Tablet  `json:"tablet"`
// 	Desktop Desktop `json:"desktop"`
// 	Mobile  Mobile  `json:"mobile"`
// }
// type Categories struct {
// 	Code string `json:"code"`
// 	Name string `json:"name"`
// }
// type Category struct {
// 	Code string `json:"code"`
// 	Name string `json:"name"`
// }
// type Subcategory struct {
// 	Code string `json:"code"`
// 	Name string `json:"name"`
// }
// type Adtech struct {
// 	Kvprogramid    string `json:"kvprogramid"`
// 	Kvvty          string `json:"kvvty"`
// 	Kvcategory     string `json:"kvcategory"`
// 	Kvsubcategory  string `json:"kvsubcategory"`
// 	Kvwebonly      bool   `json:"kvwebonly"`
// 	Kvbroadcast    bool   `json:"kvbroadcast"`
// 	Kvduration     int    `json:"kvduration"`
// 	Kvcaseprogram  string `json:"kvcaseprogram"`
// 	Kvvideoisolang string `json:"kvvideoisolang"`
// }
// type EStat struct {
// 	StreamName     string `json:"streamName"`
// 	StreamDuration int    `json:"streamDuration"`
// 	StreamGenre    string `json:"streamGenre"`
// 	Level1         string `json:"level1"`
// 	Level2         string `json:"level2"`
// 	Level3         string `json:"level3"`
// 	Level4         string `json:"level4"`
// 	Level5         string `json:"level5"`
// 	NewLevel1      string `json:"newLevel1"`
// 	NewLevel11     string `json:"newLevel11"`
// 	MediaChannel   string `json:"mediaChannel"`
// 	MediaContentID string `json:"mediaContentId"`
// 	MediaDiffMode  string `json:"mediaDiffMode"`
// }
// type Smart struct {
// 	URL string `json:"url"`
// }
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
