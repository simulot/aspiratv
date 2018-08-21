package artetv

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// cSpell:disable

// Day guide structure
// Commented out fields are kept for documentation
type guide struct {
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
	Zones []zones `json:"zones"`
}

type zones struct {
	Code code `json:"code"`
	// ContextPage    string `json:"contextPage"`
	Data []data `json:"data"`
	// DisplayOptions struct {
	// 	Layout   string      `json:"layout"`
	// 	Template string      `json:"template"`
	// 	Theme    interface{} `json:"theme"`
	// } `json:"displayOptions"`
	// Link     interface{} `json:"link"`
	// NextPage interface{} `json:"nextPage"`
	// Title    string      `json:"title"`
	// Type     string      `json:"type"`
}

type code struct {
	ID   interface{} `json:"id"`   // 0 for highlights, 1  for listing
	Name string      `json:"name"` // highlights_TV_GUIDE  listing_TV_GUIDE
}

type data struct {
	// Availability    interface{}   `json:"availability"`
	BroadcastDates []tsGuide `json:"broadcastDates"`
	// Credits         []interface{} `json:"credits"`
	Duration        seconds `json:"duration"`
	Description     string  `json:"description"`
	FullDescription string  `json:"fullDescription"`
	// Geoblocking     interface{}   `json:"geoblocking"`
	ID     string            `json:"id"`
	Images map[string]thumbs `json:"images"`
	Kind   struct {
		Code  string      `json:"code"`
		Label interface{} `json:"label"`
	} `json:"kind"`
	// LivestreamRights    interface{} `json:"livestreamRights"`
	// OfflineAvailability interface{} `json:"offlineAvailability"`
	// Partners            interface{} `json:"partners"`
	// Player              struct {
	// 	HTML        string      `json:"html"`
	// 	LinkAndroid interface{} `json:"linkAndroid"`
	// 	LinkIos     interface{} `json:"linkIos"`
	// } `json:"player"`
	ProgramID string `json:"programId"`
	// ShopURL          interface{}   `json:"shopUrl"`
	ShortDescription string `json:"shortDescription"`
	// Stickers         []interface{} `json:"stickers"`
	Subtitle string `json:"subtitle"`
	Title    string `json:"title"`
	Type     string `json:"type"`
	URL      string `json:"url"`
}

type thumbs struct {
	BlurURL     string       `json:"blurUrl"`
	Caption     string       `json:"caption"`
	Resolutions []resolution `json:"resolutions"`
}

type resolution struct {
	Height int    `json:"height"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
}

// seconds read a number of seconds and transform it into time.Duration
type seconds time.Duration

func (s *seconds) UnmarshalJSON(b []byte) error {
	if b[0] == '"' {
		b = b[1 : len(b)-1]
	}
	if string(b) == "null" {
		*s = 0
		return nil
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return fmt.Errorf("Can't parse duration in seconds: %v", err)
	}
	*s = seconds(time.Duration(i) * time.Second)
	return nil
}

func (s seconds) Duration() time.Duration { return time.Duration(s) }

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

// player structure
type player struct {
	VideoJSONPlayer struct {
		// V7T string `json:"V7T"`
		// VAD bool   `json:"VAD"` // == false ?
		// VDU int    `json:"VDU"`
		// VGB string `json:"VGB"` //  ?
		// VID string `json:"VID"` // ID
		// VPI string `json:"VPI"` // Program ID
		VRA tsAvailability `json:"VRA"` // Available from
		VRU tsAvailability `json:"VRU"` // Available to
		// VSO string `json:"VSO"` // =="replay"
		VSR map[string]streamInfo `json:"VSR"` // Video streams
		VTI string                `json:"VTI"` // Title
		// VTR string `json:"VTR"` // Show's page
		// VTU struct {
		// 	IUR string `json:"IUR"` // Show's thumbnail
		// } `json:"VTU"`
		// VTX string `json:"VTX"` // Channel ? == "ARTE"
		// VTY string `json:"VTY"` // ?? =="ARTE_NEXT"
		// VUP    string `json:"VUP"` // Show's page
		// Adtech struct {
		// 	Kvbroadcast    bool   `json:"kvbroadcast"`
		// 	Kvcaseprogram  string `json:"kvcaseprogram"`
		// 	Kvcategory     string `json:"kvcategory"`
		// 	Kvduration     int    `json:"kvduration"`
		// 	Kvprogramid    string `json:"kvprogramid"`
		// 	Kvsubcategory  string `json:"kvsubcategory"`
		// 	Kvvideoisolang string `json:"kvvideoisolang"`
		// 	Kvvty          string `json:"kvvty"`
		// 	Kvwebonly      bool   `json:"kvwebonly"`
		// } `json:"adtech"`
		// ArteClub    bool   `json:"arteClub"`
		// Autostart   bool   `json:"autostart"`
		// CaseProgram string `json:"caseProgram"` // mangled code
		Categories []codeName `json:"categories"` //Catergory
		Category   codeName   `json:"category"`   // another Catergory
		// Collections []interface{} `json:"collections"` // Empty
		// EStat       struct {
		// 	Level1         string `json:"level1"`
		// 	Level2         string `json:"level2"`
		// 	Level3         string `json:"level3"`
		// 	Level4         string `json:"level4"`
		// 	Level5         string `json:"level5"`
		// 	MediaChannel   string `json:"mediaChannel"`
		// 	MediaContentID string `json:"mediaContentId"`
		// 	MediaDiffMode  string `json:"mediaDiffMode"`
		// 	NewLevel1      string `json:"newLevel1"`
		// 	NewLevel11     string `json:"newLevel11"`
		// 	StreamDuration int    `json:"streamDuration"`
		// 	StreamGenre    string `json:"streamGenre"`
		// 	StreamName     string `json:"streamName"`
		// } `json:"eStat"`
		// EnablePreroll bool `json:"enablePreroll"`
		// Genre            string        `json:"genre"`
		// Illico           bool          `json:"illico"`
		// Kind             string        `json:"kind"`             // == SHOW
		// KindLabel        string        `json:"kindLabel"`        // == PROGRAMME
		// Language         string        `json:"language"`         // Language code like "fr"
		// LiveStreamRights bool          `json:"liveStreamRights"` // = true
		// MainPlatformCode string        `json:"mainPlatformCode"` // CREATIVE / FUTURE
		// Markings         []interface{} `json:"markings"`
		// ParentProgramID  string        `json:"parentProgramId"` // == ProgramID
		// Platform         string        `json:"platform"` // == "ARTE_NEXT"
		// Postroll         string        `json:"postroll"` // Recommendation
		ProgramID string `json:"programId"`
		// ProgramType      string        `json:"programType"` // == "BROADCAST"
		Subcategory codeName `json:"subcategory"` //Subcategory
		Subtitle    string   `json:"subtitle"`
		// TcStartFrom int      `json:"tc_start_from"` // ??
		// Tracking    struct {
		// 	Desktop struct {
		// 		EMBED string `json:"EMBED"`
		// 		WEB   string `json:"WEB"`
		// 	} `json:"desktop"`
		// 	Mobile struct {
		// 		EMBED string `json:"EMBED"`
		// 		WEB   string `json:"WEB"`
		// 	} `json:"mobile"`
		// 	Tablet struct {
		// 		EMBED string `json:"EMBED"`
		// 		WEB   string `json:"WEB"`
		// 	} `json:"tablet"`
		// } `json:"tracking"`
		VideoDurationSeconds seconds `json:"videoDurationSeconds"` // Secondes
		// VideoIsoLang         string `json:"videoIsoLang"`         // Code language iso
		// VideoPlayerURL       string `json:"videoPlayerUrl"`       // URL of this JSON
		// VideoWarning         bool   `json:"videoWarning"`         // When true, the player add a warning
	} `json:"videoJsonPlayer"`
}

type codeName struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type streamInfo struct {
	Bitrate             int    `json:"bitrate"`
	Height              int    `json:"height"`
	ID                  string `json:"id"`
	MediaType           string `json:"mediaType"`
	MimeType            string `json:"mimeType"`
	Quality             string `json:"quality"`
	URL                 string `json:"url"`
	VersionCode         string `json:"versionCode"`
	VersionLibelle      string `json:"versionLibelle"`
	VersionProg         int    `json:"versionProg"`
	VersionShortLibelle string `json:"versionShortLibelle"`
	Width               int    `json:"width"`
}

type searchResults struct {
	Code struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"code"`
	ContextPage interface{} `json:"contextPage"`
	// Data []struct {
	// 	// Availability  interface{} `json:"availability"`
	// 	// ChildrenCount interface{} `json:"childrenCount"`
	// 	Description string  `json:"description"`
	// 	Duration    seconds `json:"duration"`
	// 	// Geoblocking   struct {
	// 	// 	Code  string `json:"code"`
	// 	// 	Label string `json:"label"`
	// 	// } `json:"geoblocking"`
	// 	// ID     string `json:"id"`
	// 	Images map[string]thumbs `json:"images"`
	// 	// Images struct {
	// 	// 	Banner    interface{} `json:"banner"`
	// 	// 	Landscape struct {
	// 	// 		BlurURL     string `json:"blurUrl"`
	// 	// 		Caption     string `json:"caption"`
	// 	// 		Resolutions []struct {
	// 	// 			Height int    `json:"height"`
	// 	// 			URL    string `json:"url"`
	// 	// 			Width  int    `json:"width"`
	// 	// 		} `json:"resolutions"`
	// 	// 	} `json:"landscape"`
	// 	// 	Portrait interface{} `json:"portrait"`
	// 	// 	Square   interface{} `json:"square"`
	// 	// } `json:"images"`
	// 	Kind struct {
	// 		Code  string `json:"code"`
	// 		Label string `json:"label"`
	// 	} `json:"kind"`
	// 	ProgramID string `json:"programId"`
	// 	// Stickers  []interface{} `json:"stickers"`
	// 	Subtitle string `json:"subtitle"`
	// 	Title    string `json:"title"`
	// 	Type     string `json:"type"`
	// 	URL      string `json:"url"`
	// } `json:"data"`
	Data []data `json:"data"`
	// DisplayOptions struct {
	// 	Layout   string      `json:"layout"`
	// 	Template string      `json:"template"`
	// 	Theme    interface{} `json:"theme"`
	// } `json:"displayOptions"`
	Link struct {
		// Page  string `json:"page"`
		Title string `json:"title"`
		// URL   string `json:"url"`
	} `json:"link"`
	NextPage string `json:"nextPage"`
	// Title    string      `json:"title"`
	// Type     interface{} `json:"type"`
}
