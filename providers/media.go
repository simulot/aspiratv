package providers

import (
	"github.com/simulot/aspiratv/metadata/nfo"
	"github.com/simulot/aspiratv/providers/matcher"
)

// MetaDataHandler represents a struct for managing media's metadata
type MetaDataHandler interface {
	GetMediaInfo() *nfo.MediaInfo               // return a pointer to MediaInfo struct
	GetMediaPath(showPath string) string        // Returns mp4 path as showPath/.../file.mp4
	GetMediaPathMatcher(showPath string) string // Returns a name matcher for mis numbered episodes
	GetNFOPath(showPath string) string          // Returns nfo path as showPath/.../file.nfo
	WriteNFO(showPath string) error             // Write nfo file showPath/.../file.nfo
	Accepted(m *matcher.MatchRequest) bool      // TODO check if this is the right place ofr this
}

// ShowType says if the media is a movie (one time broadcast), TVShows (recurring show) or a series (with seasons and episodes)
type ShowType int

// ShowType values
const (
	Series ShowType = iota // Series has seasons and episodes
	Movie                  // Just one media
	// TVShow                 // Regular TV show TBD
)

// Media represents a media to be handled.
type Media struct {
	ID       string                // Show ID
	ShowType ShowType              // Movie or Series?
	Metadata MetaDataHandler       // Carry metadata scrapped online
	Match    *matcher.MatchRequest // Matched request
	ShowPath string                // Path of the show/media
}

func (m *Media) SetMetaData(info MetaDataHandler) {
	m.Metadata = info
}
