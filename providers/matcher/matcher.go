package matcher

import (
	"regexp"
)

// MatchRequest holds criterions for selecting show
type MatchRequest struct {
	// Fields for matching
	Show        string
	ShowID      string // Future use
	Title       string
	TitleID     string // Future use
	Pitch       string
	Provider    string
	Playlist    string // Playlist search is implemented in providers.
	MaxAgedDays int    // Retrive media younger than MaxAgedDays when not zero

	Destination   string // Destination name when found
	RetentionDays int    // Media retention time, when not zero the system will delete old files
	TitleFilter   Filter // ShowTitle or Episode title must match this regexp to be downloaded
	TitleExclude  Filter // ShowTitle and Episode title must not match this regexp to be downloaded
}

type Filter struct {
	*regexp.Regexp
}

// Handle Filter as string for JSON configuration

func (t Filter) MarshalJSON() ([]byte, error) {

	return []byte(`"` + t.String() + `"`), nil
}

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
