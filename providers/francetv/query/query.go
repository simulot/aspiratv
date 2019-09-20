package query

type QueryResults struct {
	Results []Results `json:"results"`
}
type Dates struct {
	FirstPublicationDate  int         `json:"first_publication_date"`
	LastPublicationDate   int         `json:"last_publication_date"`
	LastUnpublicationDate interface{} `json:"last_unpublication_date"`
	BroadcastBeginDate    int         `json:"broadcast_begin_date"`
	BroadcastEndDate      int         `json:"broadcast_end_date"`
}
type Free struct {
	BeginDate int `json:"begin_date"`
	EndDate   int `json:"end_date"`
}
type Orange struct {
	BeginDate int `json:"begin_date"`
	EndDate   int `json:"end_date"`
}
type Web struct {
	BeginDate int `json:"begin_date"`
	EndDate   int `json:"end_date"`
}
type Replay struct {
	Free   Free   `json:"free"`
	Orange Orange `json:"orange"`
	Web    Web    `json:"web"`
}
type Ranges struct {
	Replay Replay `json:"replay"`
}
type Urls struct {
	W400  string `json:"w:400"`
	W800  string `json:"w:800"`
	W1024 string `json:"w:1024"`
}
type Vignette16X9 struct {
	OriginalPath string `json:"original_path"`
	OriginalName string `json:"original_name"`
	Urls         Urls   `json:"urls"`
}
type Urls struct {
	W265 string `json:"w:265"`
	W300 string `json:"w:300"`
	W400 string `json:"w:400"`
}
type Carre struct {
	OriginalPath string `json:"original_path"`
	OriginalName string `json:"original_name"`
	Urls         Urls   `json:"urls"`
}
type Formats struct {
	Vignette16X9 Vignette16X9 `json:"vignette_16x9"`
	Carre        Carre        `json:"carre"`
}
type Image struct {
	ID      int     `json:"id"`
	Title   string  `json:"title"`
	Credit  string  `json:"credit"`
	Formats Formats `json:"formats"`
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
	ID           int         `json:"id"`
	Class        string      `json:"class"`
	Type         string      `json:"type"`
	Label        string      `json:"label"`
	URL          string      `json:"url"`
	URLComplete  string      `json:"url_complete"`
	Season       interface{} `json:"season"`
	EpisodeCount interface{} `json:"episode_count"`
}
type Logo struct {
	OriginalName string `json:"original_name"`
	OriginalPath string `json:"original_path"`
	ID           int    `json:"id"`
}
type Program struct {
	ID           int         `json:"id"`
	Class        string      `json:"class"`
	Type         string      `json:"type"`
	Label        string      `json:"label"`
	URL          string      `json:"url"`
	URLComplete  string      `json:"url_complete"`
	Season       interface{} `json:"season"`
	EpisodeCount interface{} `json:"episode_count"`
	Logo         Logo        `json:"logo"`
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
type Title struct {
	Value        string        `json:"value"`
	MatchLevel   string        `json:"matchLevel"`
	MatchedWords []interface{} `json:"matchedWords"`
}
type Description struct {
	Value            string   `json:"value"`
	MatchLevel       string   `json:"matchLevel"`
	FullyHighlighted bool     `json:"fullyHighlighted"`
	MatchedWords     []string `json:"matchedWords"`
}
type Label struct {
	Value        string        `json:"value"`
	MatchLevel   string        `json:"matchLevel"`
	MatchedWords []interface{} `json:"matchedWords"`
}
type Categories struct {
	Label Label `json:"label"`
}
type Type struct {
	Value        string        `json:"value"`
	MatchLevel   string        `json:"matchLevel"`
	MatchedWords []interface{} `json:"matchedWords"`
}
type Label struct {
	Value            string   `json:"value"`
	MatchLevel       string   `json:"matchLevel"`
	FullyHighlighted bool     `json:"fullyHighlighted"`
	MatchedWords     []string `json:"matchedWords"`
}
type Program struct {
	Type  Type  `json:"type"`
	Label Label `json:"label"`
}
type HighlightResult struct {
	Title       Title        `json:"title"`
	Description Description  `json:"description"`
	Categories  []Categories `json:"categories"`
	Program     Program      `json:"program"`
}
type Hits struct {
	ID                      int             `json:"id"`
	Class                   string          `json:"class"`
	Type                    string          `json:"type"`
	Title                   string          `json:"title"`
	HeadlineTitle           string          `json:"headline_title"`
	Description             string          `json:"description"`
	Text                    string          `json:"text"`
	URLPage                 string          `json:"url_page"`
	Path                    string          `json:"path"`
	Duration                int             `json:"duration"`
	SeasonNumber            int             `json:"season_number"`
	EpisodeNumber           int             `json:"episode_number"`
	IsAudioDescripted       bool            `json:"is_audio_descripted"`
	IsPreviouslyBroadcasted bool            `json:"is_previously_broadcasted"`
	IsMultiLingual          bool            `json:"is_multi_lingual"`
	IsSubtitled             bool            `json:"is_subtitled"`
	IsPreview               bool            `json:"is_preview"`
	IsSponsored             bool            `json:"is_sponsored"`
	Director                string          `json:"director"`
	Producer                string          `json:"producer"`
	Presenter               string          `json:"presenter"`
	Casting                 string          `json:"casting"`
	Characters              string          `json:"characters"`
	ProductionYear          int             `json:"production_year"`
	PressRating             interface{}     `json:"press_rating"`
	PeopleRating            interface{}     `json:"people_rating"`
	LabelSvod               interface{}     `json:"label_svod"`
	Favorite                interface{}     `json:"favorite"`
	AdsBlocked              bool            `json:"ads_blocked"`
	LiveAdsBlocked          bool            `json:"live_ads_blocked"`
	Dates                   Dates           `json:"dates"`
	Ranges                  Ranges          `json:"ranges"`
	Image                   Image           `json:"image,omitempty"`
	Categories              []Categories    `json:"categories"`
	Channels                []Channels      `json:"channels"`
	Program                 Program         `json:"program"`
	Season                  Season          `json:"season"`
	RatingCsaCode           string          `json:"rating_csa_code"`
	SiID                    string          `json:"si_id"`
	FreeID                  int             `json:"free_id"`
	OrangeID                string          `json:"orange_id"`
	ObjectID                string          `json:"objectID"`
	HighlightResult         HighlightResult `json:"_highlightResult"`
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
