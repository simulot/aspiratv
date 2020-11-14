package nfo

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"

	"github.com/simulot/aspiratv/providers/matcher"
)

// Movie holds metadata for movies
type Movie struct {
	XMLName xml.Name `xml:"movie"`
	MediaInfo
}

// Accepted check if Title matches the filter
func (n Movie) Accepted(m *matcher.MatchRequest) bool {
	if m.TitleExclude.Regexp != nil {
		if m.TitleExclude.Regexp.MatchString(n.Title) {
			return false
		}
	}
	if m.TitleFilter.Regexp != nil {
		if m.TitleFilter.Regexp.MatchString(n.Title) {
			return true
		}
		return false
	}
	return true
}

// GetMediaInfo return a pointer to MediaInfo struct
func (n *Movie) GetMediaInfo() *MediaInfo {
	return &n.MediaInfo
}

// GetNFOPath give the path where the episode's NFO should be
func (n Movie) GetNFOPath(destination string) string {
	cleanTitle := FileNameCleaner(n.Title)
	return filepath.Join(destination, FileNameCleaner(n.Title), cleanTitle+".nfo")
}

// GetSeasonNFOPath returns the path for TVShow.nfo
func (n Movie) GetSeasonNFOPath(destination string) string {
	return ""
}

// GetShowNFOPath returns the path for TVShow.nfo
func (n Movie) GetShowNFOPath(destination string) string {
	return ""
}

// GetMediaPath returns the media path
func (n Movie) GetMediaPath(destination string) string {
	cleanTitle := FileNameCleaner(n.Title)
	return filepath.Join(destination, cleanTitle, cleanTitle+".mp4")
}

// GetSeriesPath gives path for the whole series
func (n Movie) GetSeriesPath(destination string) string {
	return destination
}

// GetSeasonPath give the path for the series' season
func (n Movie) GetSeasonPath(destination string) string {
	return n.GetSeriesPath(destination)
}

func (n Movie) GetMediaPathMatcher(destination string) string {
	return n.GetMediaPath(destination)
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
