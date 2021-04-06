package nfo

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
)

// EpisodeDetails gives details of a given episode
type EpisodeDetails struct {
	XMLName xml.Name `xml:"episodedetails"`
	MediaInfo
	// TVShow *TVShow `xml:"-"`
}

// GetMediaInfo return a pointer to MediaInfo struct
func (n *EpisodeDetails) GetMediaInfo() *MediaInfo {
	return &n.MediaInfo
}

// WriteNFO file at expected place
// TODO remove destination and get it from show path
func (n *EpisodeDetails) WriteNFO(destination string) error {
	err := os.MkdirAll(filepath.Dir(destination), 0777)
	if err != nil {
		return fmt.Errorf("Can't create %s :%w", destination, err)
	}
	f, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("Can't create tvshow.nfo :%w", err)
	}
	defer f.Close()

	_, err = f.WriteString(xml.Header)
	err = xml.NewEncoder(f).Encode(n)
	if err != nil {
		return fmt.Errorf("Can't encode tvshow.nfo :%w", err)
	}

	return nil
}
