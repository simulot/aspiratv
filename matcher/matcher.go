package matcher

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/simulot/aspiratv/metadata/nfo"
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

	// Fields for managing download
	Type               nfo.ShowType   // Determine the type of show for file layout and naming
	Destination        string         // Download base destination code defined in config
	ShowRootPath       string         `json:"ShowPath"` // Show/Movie path. Bypass destinations. For expisodes, actual season will append to the path
	SeasonPathTemplate TemplateString // Template for season path, can be empty to skip season in path. When missing uses default naming
	ShowNameTemplate   TemplateString // Template for the name of mp4 file, can't be empty. When missing, uses default naming
	RetentionDays      int            // Media retention time, when not zero the system will delete old files
	TitleFilter        Filter         // ShowTitle or Episode title must match this regexp to be downloaded
	TitleExclude       Filter         // ShowTitle and Episode title must not match this regexp to be downloaded
	KeepBonus          bool           // When trie bonuses and trailer are retrieved
	Force              bool           // True to force  medias

}

func (m MatchRequest) Validate(destinations map[string]string) error {
	if m.Show == "" {
		return errors.New("Missing show name")
	}

	if m.Provider == "" {
		return errors.New("Missing provider")
	}

	if m.ShowRootPath == "" && m.Destination == "" {
		return errors.New("Missing show path or destination")
	}

	if m.ShowRootPath != "" {
		p := os.ExpandEnv(m.ShowRootPath)
		p, err := filepath.Abs(p)
		if err != nil {
			return err
		}
		err = os.MkdirAll(p, 0755)
		if err != nil {
			return err
		}
		m.ShowRootPath = p
	}

	return nil
}

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

// Filter is a wrapper for regexp
// Implement JSON and Value interfaces
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
		return t.Set(string(b))
	}
	return nil
}

// Set value
func (t *Filter) Set(s string) error {
	re, err := regexp.Compile(s)
	if err != nil {
		return err
	}
	t.Regexp = re
	return nil
}

func (t Filter) String() string {
	if t.Regexp != nil {
		return t.Regexp.String()
	}
	return ""
}

// Type used in command line
func (Filter) Type() string { return "regexp-filter" }

type TemplateString struct {
	S string
	T *template.Template
}

func (t TemplateString) String() string {
	return t.S
}
func (t TemplateString) Type() string {
	return "name-template"
}

func (t *TemplateString) Set(s string) error {
	var err error
	t.T, err = template.New("").Parse(s)
	if err != nil {
		return err
	}
	t.S = s
	return nil

}

//MarshalJSON returns a  string from regexp and place it in the JSON stream
func (t TemplateString) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.S + `"`), nil
}

// UnmarshalJSON takes the string from the stream and compile the template
func (t *TemplateString) UnmarshalJSON(b []byte) error {
	var err error
	// if b[0] == '"' {
	// 	b = b[1 : len(b)-1]
	// }
	err = json.Unmarshal(b, &t.S)
	if err != nil {
		return err
	}

	t.T, err = template.New("").Parse(t.S)
	if err != nil {
		return err
	}
	return nil
}

// Accepted check if ShowTitle or episode Title matches the filter
func (m *MatchRequest) Accepted(n *nfo.MediaInfo) bool {
	if m.TitleExclude.Regexp != nil {
		if m.TitleExclude.Regexp.MatchString(n.Showtitle) {
			return false
		}
		if m.TitleExclude.Regexp.MatchString(n.Title) {
			return false
		}
	}
	if m.TitleFilter.Regexp != nil {
		if m.TitleFilter.Regexp.MatchString(n.Showtitle) {
			return true
		}
		if m.TitleFilter.Regexp.MatchString(n.Title) {
			return true
		}
		return false
	}
	return true
}
