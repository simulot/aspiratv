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
	ID         string    `json:"id,omitempty"`         // Unique ID given by the program
	Title      string    `json:"title,omitempty"`      // Episode title, show title, movie title
	Show       string    `json:"show,omitempty"`       // Name of the show
	Type       MediaType `json:"type,omitempty"`       // Series, TVShow, Movie
	Season     int       `json:"season,omitempty"`     // Season number
	Episode    int       `json:"episode,omitempty"`    // Episode number
	Aired      time.Time `json:"aired,omitempty"`      // date of the 1st broadcast
	Year       int       `json:"year,omitempty"`       // Year of broadcast
	Plot       string    `json:"plot,omitempty"`       // Show plot
	Actors     []Person  `json:"actors,omitempty"`     // Show actors
	Team       []Person  `json:"team,omitempty"`       // Production team
	Categories []string  `json:"categories,omitempty"` //
	Thumbs     []Image   `json:"thumbs,omitempty"`     //
	Provider   string    `json:"provider,omitempty"`   // Provider where the show has been taken
	Channel    string    `json:"channel,omitempty"`    // TV Channel
	PageURL    string    `json:"page_url,omitempty"`   // Episode page on web site
	IsBonus    bool      `json:"is_bonus,omitempty"`   //

	SeasonInfo *SeasonInfo `json:"-"` // Season metadata
	ShowInfo   *ShowInfo   `json:"-"` // Show metadata

	StreamURL string `json:"-"`
}

type SeasonInfo struct {
	ID      string  `json:"id,omitempty"`
	Title   string  `json:"title,omitempty"`  // Season title
	Plot    string  `json:"plot,omitempty"`   //Season plot
	Season  int     `json:"season,omitempty"` // Season Number
	Year    int     `json:"year,omitempty"`   // Broadcasting year
	PageURL string  `json:"page_url,omitempty"`
	Thumbs  []Image `json:"thumbs,omitempty"` //

	// Episods []*Media
}

type ShowInfo struct {
	ID      string    `json:"id,omitempty"`
	Title   string    `json:"title,omitempty"` // Show name
	Plot    string    `json:"plot,omitempty"`  // Show plot
	Type    MediaType `json:"type,omitempty"`
	PageURL string    `json:"page_url,omitempty"`
	Thumbs  []Image   `json:"thumbs,omitempty"`
}

type Person struct {
	FullName string
	Role     string
	Thumbs   []Image
}

type Image struct {
	ID     string
	Aspect string // poster, banner, portrait
	URL    string
}
