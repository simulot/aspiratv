package nfo

import (
	"fmt"
	"time"
)

// ID for Shows / Episode
type ID struct {
	ID      string `xml:",chardata"`
	Type    string `xml:"type,attr,omitempty"`
	Default string `xml:"default,attr,omitempty"`
}

// Actor describe actors
type Actor struct {
	Name  string `xml:"name,omitempty"`
	Role  string `xml:"role,omitempty"`
	Order string `xml:"order,omitempty"`
	Type  string `xml:"type,omitempty"`
	Thumb string `xml:"thumb,omitempty"`
}

// ThumbLevel is the level where the thumbnail belongs to
type ThumbLevel int

// Thumb record
type Thumb struct {
	Aspect  string `xml:"aspect,attr,omitempty"`
	Preview string `xml:"preview,attr,omitempty"`
	URL     string `xml:",chardata"`
}

// MediaInfo is the shared part of metadata
type MediaInfo struct {
	Title          string   `xml:"title,omitempty"`
	Showtitle      string   `xml:"showtitle,omitempty"`
	Season         int      `xml:"season,omitempty"`
	Episode        int      `xml:"episode,omitempty"`
	DisplaySeason  int      `xml:"displayseason,omitempty"`
	DisplayEpisode int      `xml:"displayepisode,omitempty"`
	Plot           string   `xml:"plot,omitempty"`
	Thumb          []Thumb  `xml:"-"`
	UniqueID       []ID     `xml:"uniqueid,omitempty"`
	Genre          []string `xml:"genre,omitempty"`
	Credits        []string `xml:"credits,omitempty"`
	Director       []string `xml:"director,omitempty"`
	Aired          Aired    `xml:"aired,omitempty"`
	Studio         string   `xml:"studio,omitempty"`
	Actor          []Actor  `xml:"actor,omitempty"`
	Tag            []string `xml:"tag,omitempty"`

	URL        string  `xml:"-"` // Media URL
	IsSpecial  bool    `xml:"-"` // True when special episode
	SeasonInfo *Season `xml:"-"` // Possible Season nfo
	TVShow     *TVShow `xml:"-"` // Possible TVShow nfo
}

// Aired type helper
type Aired time.Time

// UnmarshalText grab aired field and turn it into time
func (f *Aired) UnmarshalText(text []byte) error {
	t, err := time.Parse("2006-01-02", string(text))
	if err != nil {
		return fmt.Errorf("Can't parse Aired. %w", err)
	}
	*f = Aired(t)
	return nil
}

// Time convert Aired to Time
func (f Aired) Time() time.Time {
	return time.Time(f)
}

// MarshalText write Aired correctly
func (f Aired) MarshalText() ([]byte, error) {
	return []byte(f.Time().Format("2006-01-02")), nil
}
