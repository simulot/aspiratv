package models

import "time"

type MediaType int

//go:generate stringer -type=MediaType
const (
	TypeUnknown    MediaType = iota
	TypeCollection           // We don't know yet the actual type, but this is a collection, Arte?
	TypeSeries               // Series with seasons and episodes
	TypeTVShow               // TV Show or magazine
	TypeMovie                // Movie
)

type MediaInfo struct {
	ID         string    // Unique ID given by the program
	Title      string    // Episode title, show title, movie title
	Type       MediaType // Series, TVShow, Movie
	Season     int       // Season number
	Episode    int       // Episode number
	Aired      time.Time // date of the 1st broadcast
	Year       int       // Year of broadcast
	Show       string    // Name of the show
	Plot       string    // Show plot
	Actors     []Person  // Show actors
	Team       []Person  // Production team
	Categories []string  //
	Thumbs     []Image   //
	Provider   string    // Provider where the show has been taken
	Channel    string    // TV Channel
	PageURL    string    // Episode page on web site

	SeasonInfo *SeasonInfo // Season metadata
	ShowInfo   *ShowInfo   // Show metadata

	StreamURL string
}

type SeasonInfo struct {
	ID      string
	Title   string // Season title
	Plot    string //Season plot
	Season  int    // Season Number
	Year    int    // Broadcasting year
	PageURL string
	Thumbs  []Image //

	// Episods []*Media
}

type ShowInfo struct {
	ID      string
	Title   string // Show name
	Plot    string // Show plot
	Type    MediaType
	PageURL string
	Thumbs  []Image
	// Seasons []*SeasonInfo
	// Episods []*Media
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
