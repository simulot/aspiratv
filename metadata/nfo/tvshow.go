package nfo

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
)

// TVShow description named tvshow.nfo
type TVShow struct {
	XMLName       xml.Name `xml:"tvshow"`
	Title         string   `xml:"title,omitempty"`
	OriginalTitle string   `xml:"originaltitle,omitempty"`
	Plot          string   `xml:"plot,omitempty"`
	Userrating    string   `xml:"userrating,omitempty"`
	MPAA          string   `xml:"mpaa,omitempty"`
	UniqueID      []ID     `xml:"uniqueid,omitempty"`
	Genre         []string `xml:"genre,omitempty"`
	Studio        string   `xml:"studio,omitempty"`
	Actor         []Actor  `xml:"actor,omitempty"`
	Thumb         []Thumb  `xml:"-"`
}

// GetNFOPath returns the path for TVShow.nfo
func (n TVShow) GetNFOPath(destination string) string {
	return filepath.Join(destination, FileNameCleaner(n.Title), "tvshow.nfo")
}

// WriteNFO TVShow's NFO and download thumbnails
func (n *TVShow) WriteNFO(destination string) error {
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
