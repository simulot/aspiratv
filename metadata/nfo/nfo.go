package nfo

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/simulot/aspiratv/providers"
)

// ID for Shows / Episode
type ID struct {
	ID      string `xml:",chardata"`
	Type    string `xml:"type,attr"`
	Default string `xml:"default,attr"`
}

// Actor describe actors
type Actor struct {
	Name  string `xml:"name"`
	Role  string `xml:"role"`
	Order string `xml:"order"`
	Thumb string `xml:"thumb"`
}

// Thumb record
type Thumb struct {
	Aspect  string `xml:"aspect,attr"`
	Preview string `xml:"preview,attr"`
	Path    string `xml:",chardata"`
}

// TVShow description named tvshow.nfo
type TVShow struct {
	XMLName       xml.Name `xml:"tvshow"`
	Title         string   `xml:"title"`
	OriginalTitle string   `xml:"originaltitle"`
	Plot          string   `xml:"plot"`
	Userrating    string   `xml:"userrating"`
	MPAA          string   `xml:"mpaa"`
	UniqueID      []ID     `xml:"uniqueid"`
	Genre         []string `xml:"genre"`
	Studio        string   `xml:"studio"`
	Actor         []Actor  `xml:"actor"`
	Thumb         []Thumb  `xml:"thumb"`
}

// EpisodeDetails gives details of a given episode
type EpisodeDetails struct {
	XMLName        xml.Name `xml:"episodedetails"`
	Text           string   `xml:",chardata"`
	Title          string   `xml:"title"`
	Showtitle      string   `xml:"showtitle"`
	Season         string   `xml:"season"`
	Episode        string   `xml:"episode"`
	DisplaySeason  string   `xml:"displayseason"`
	DisplayEpisode string   `xml:"displayepisode"`
	Plot           string   `xml:"plot"`
	Thumb          []Thumb  `xml:"thumb"`
	UniqueID       []ID     `xml:"uniqueid"`
	Genre          []string `xml:"genre"`
	Credits        []string `xml:"credits"`
	Director       []string `xml:"director"`
	Aired          string   `xml:"aired"`
	Studio         string   `xml:"studio"`
	Actor          []Actor  `xml:"actor"`
}

// ShowNFOPath give the show's NFO path
func ShowNFOPath(s *providers.Show) string {
	return filepath.Join(providers.GetShowPath(s), "tvshow.nfo")
}

// WriteShowData write show's NFO file
func WriteShowData(s *providers.Show) error {
	d := TVShow{
		Title: s.Show,
		Plot:  s.ShowPitch,
		UniqueID: []ID{
			{
				Type: s.Channel,
				ID:   s.ID,
			},
		},
	}

	f, err := os.Create(ShowNFOPath(s))
	if err != nil {
		return fmt.Errorf("Can't write tvshow.nfo: %w", err)
	}
	defer f.Close()

	err = xml.NewEncoder(f).Encode(&d)
	if err != nil {
		return fmt.Errorf("Can't encode tvshow.nfo: %w", err)
	}
	return nil
}

func EpisodeNFOPath(s *providers.Show) string {
	return strings.TrimSuffix(providers.GetShowPath(s), ".mp4") + ".nfo"
}
func WriteEpisodeData(s *providers.Show) error {
	d := EpisodeDetails{
		Title:   s.Title,
		Season:  s.Season,
		Episode: s.Episode,
		Plot:    s.Pitch,
		Thumb: []Thumb{
			{
				Aspect:  "landscape",
				Preview: "",
				Path:    s.ThumbnailURL,
			},
		},
	}
	f, err := os.Create(EpisodeNFOPath(s))
	if err != nil {
		return fmt.Errorf("Can't write episode's .nfo: %w", err)
	}
	defer f.Close()

	err = xml.NewEncoder(f).Encode(&d)
	if err != nil {
		return fmt.Errorf("Can't encode episode's .nfo: %w", err)
	}
	return nil
}
