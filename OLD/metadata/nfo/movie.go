package nfo

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
)

// Movie holds metadata for movies
type Movie struct {
	XMLName xml.Name `xml:"movie"`
	MediaInfo
}

// GetMediaInfo return a pointer to MediaInfo struct
func (n *Movie) GetMediaInfo() *MediaInfo {
	return &n.MediaInfo
}

// WriteNFO file at expected place
func (n *Movie) WriteNFO(destination string) error {
	err := os.MkdirAll(filepath.Dir(destination), 0777)
	if err != nil {
		return fmt.Errorf("Can't create %s :%w", destination, err)
	}

	f, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("Can't create %s :%w", destination, err)
	}
	defer f.Close()
	_, err = f.WriteString(xml.Header)
	if err != nil {
		return fmt.Errorf("Can't encode %s :%w", destination, err)
	}
	err = xml.NewEncoder(f).Encode(n)
	if err != nil {
		return fmt.Errorf("Can't encode %s :%w", destination, err)
	}
	return nil
}
