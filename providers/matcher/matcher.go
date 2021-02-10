package matcher

import (
	"regexp"
	"strings"
)

// MatchRequest holds criterions for selecting show
type MatchRequest struct {
	Provider string
	// Fields for matching
	Show  string
	Title string
	// ShowID   string // Future use
	// TitleID  string // Future use
	// Pitch    string // Future use
	// Playlist    string // Playlist search is implemented in providers.
	MaxAgedDays int // Retrive media younger than MaxAgedDays when not zero

	Destination   string // Destination name when found
	ShowRootPath  string `json:"ShowPath"` // Show/Movie path. For expisodes, actual season will append to the path
	RetentionDays int    // Media retention time, when not zero the system will delete old files
	TitleFilter   Filter // ShowTitle or Episode title must match this regexp to be downloaded
	TitleExclude  Filter // ShowTitle and Episode title must not match this regexp to be downloaded
	KeepBonus     bool   // When trie bonuses and trailer are retrieved
}

// TODO implement IsTitleMatch for all providers
func (m MatchRequest) IsTitleMatch(title string) bool {
	if m.TitleExclude.Regexp != nil {
		if m.TitleExclude.Regexp.MatchString(title) {
			return false
		}
	}
	if m.TitleFilter.Regexp != nil {
		if m.TitleFilter.Regexp.MatchString(title) {
			return true
		}
	}
	title = strings.ToUpper(strings.TrimSpace(title))
	return strings.Contains(title, m.Title)
}

// // TODO implement IsShowMatch for all providers
// func (m MatchRequest) IsShowMatch(showTitle string) bool {
// 	if m.TitleExclude.Regexp != nil {
// 		if m.TitleExclude.Regexp.MatchString(showTitle) {
// 			return false
// 		}
// 	}
// 	if m.TitleFilter.Regexp != nil {
// 		if m.TitleFilter.Regexp.MatchString(showTitle) {
// 			return true
// 		}
// 	}
// 	showTitle = strings.ToLower(strings.TrimSpace(showTitle))
// 	return strings.Contains(showTitle, m.Show)
// }

// Filter is a wrapper for regexp
type Filter struct {
	*regexp.Regexp
}

//MarshalJSON returns a  string from regexp and place it in the JSON stream
func (t Filter) MarshalJSON() ([]byte, error) {

	return []byte(`"` + t.String() + `"`), nil
}

// UnmarshalJSON takes the string from the stream and compile the regexp
func (t *Filter) UnmarshalJSON(b []byte) error {
	if b[0] == '"' {
		b = b[1 : len(b)-1]
	}
	t.Regexp = nil
	if len(b) > 0 {
		re, err := regexp.Compile(string(b))
		if err != nil {
			return err
		}
		t.Regexp = re
	}
	return nil
}
