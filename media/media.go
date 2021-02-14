package media

import (
	"github.com/simulot/aspiratv/matcher"
	"github.com/simulot/aspiratv/metadata/nfo"
)

// MetaDataHandler represents a struct for managing media's metadata
type MetaDataHandler interface {
	GetMediaInfo() *nfo.MediaInfo // return a pointer to MediaInfo struct
	// GetMediaPath(showPath string) string        // Returns mp4 path as showPath/.../file.mp4
	// GetMediaPathMatcher(showPath string) string // Returns a name matcher for mis numbered episodes
	// GetNFOPath(showPath string) string     // Returns nfo path as showPath/.../file.nfo
	WriteNFO(showPath string) error // Write nfo file showPath/.../file.nfo
	// Accepted(m *matcher.MatchRequest) bool // TODO check if this is the right place ofr this
}

// Media represents a media to be handled.
type Media struct {
	ID       string                // Show ID
	Metadata MetaDataHandler       // Carry metadata scrapped online
	Match    *matcher.MatchRequest // Matched request
	ShowPath string                // Path of the show/media
}

func (m *Media) SetMetaData(info MetaDataHandler) {
	m.Metadata = info
}
