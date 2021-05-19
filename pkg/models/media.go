package models

import (
	"time"
)

type MediaType int

//go:generate enumer -type=MediaType -json
const (
	TypeUnknown    MediaType = iota
	TypeCollection           // We don't know yet the actual type, but this is a collection, Arte?
	TypeSeries               // Series with seasons and episodes
	TypeTVShow               // TV Show or magazine
	TypeMovie                // Movie
)

type MediaInfo struct {
	ID       string    `json:"id,omitempty"`       // Unique ID given by the program
	Title    string    `json:"title,omitempty"`    // Episode title, show title, movie title
	Show     string    `json:"show,omitempty"`     // Name of the show
	Type     MediaType `json:"type,omitempty"`     // Series, TVShow, Movie
	Season   int       `json:"season,omitempty"`   // Season number
	Episode  int       `json:"episode,omitempty"`  // Episode number
	Aired    time.Time `json:"aired,omitempty"`    // date of the 1st broadcast
	Year     int       `json:"year,omitempty"`     // Year of broadcast
	Plot     string    `json:"plot,omitempty"`     // Show plot
	Actors   []Person  `json:"actors,omitempty"`   // Show actors
	Images   []Image   `json:"images,omitempty"`   //
	Provider string    `json:"provider,omitempty"` // Provider where the show has been taken
	Channel  string    `json:"channel,omitempty"`  // TV Channel
	PageURL  string    `json:"page_url,omitempty"` // Episode page on web site
	IsBonus  bool      `json:"is_bonus,omitempty"` //
	Credits  []string  `json:"credits,omitempty"`
	Tags     []string  `json:"tags,omitempty"`

	SeasonInfo *SeasonInfo `json:"-,omitempty"` // Season metadata
	ShowInfo   *ShowInfo   `json:"-,omitempty"` // Show metadata

	StreamURL string `json:"-,omitempty"` // Actual url for downloading the video, may be transient
}

type SeasonInfo struct {
	ID      string  `json:"id,omitempty"`
	Title   string  `json:"title,omitempty"`  // Season title
	Plot    string  `json:"plot,omitempty"`   //Season plot
	Season  int     `json:"season,omitempty"` // Season Number
	Year    int     `json:"year,omitempty"`   // Broadcasting year
	PageURL string  `json:"page_url,omitempty"`
	Images  []Image `json:"images,omitempty"` //

	// Episods []*Media
}

type ShowInfo struct {
	ID      string    `json:"id,omitempty"`
	Title   string    `json:"title,omitempty"` // Show name
	Plot    string    `json:"plot,omitempty"`  // Show plot
	Type    MediaType `json:"type,omitempty"`
	PageURL string    `json:"page_url,omitempty"`
	Images  []Image   `json:"images,omitempty"`
}

type Person struct {
	FullName string
	Role     string
	Thumbs   []Image
}

type Image struct {
	ID     string
	Aspect string // poster, backdrop, portrait
	URL    string
}
