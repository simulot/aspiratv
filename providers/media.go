package providers

import "github.com/simulot/aspiratv/metadata/nfo"

// MetaDataHandler represents a struct for managing media's metadata
type MetaDataHandler interface {
	GetMediaInfo() *nfo.MediaInfo
	GetNFOPath(destination string) string
	GetSeasonNFOPath(destination string) string
	GetShowNFOPath(destination string) string
	GetSeriesPath(destination string) string
	GetMediaPath(destination string) string
	GetSeasonPath(destination string) string
	GetMediaPathMatcher(destination string) string
	WriteNFO(destination string) error
}

// ShowType says if the media is a movie (one time broadcast), TVShows (recurring show) or a series (with seasons and episodes)
type ShowType int

// ShowType values
const (
	Series ShowType = iota // Series has seasons and episodes
	Movie                  // Just one media

)

// Media represents a media to be handled.
type Media struct {
	ID       string          // Show ID
	ShowType ShowType        // Movie or Series?
	Metadata MetaDataHandler // Carry metadata scrapped online
	Match    *MatchRequest   // Matched request
}

func (m *Media) SetMetaData(info MetaDataHandler) {
	m.Metadata = info
}
