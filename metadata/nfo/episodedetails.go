package nfo

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// GetSeriesPath gives path for the whole series
func (n EpisodeDetails) GetSeriesPath(destination string) string {
	return filepath.Join(destination, FileNameCleaner(n.Showtitle))
}

// GetSeasonPath give the path for the series' season
func (n *EpisodeDetails) GetSeasonPath(destination string) string {
	season := "Season "
	if n.Season <= 0 {
		season += "00"
	} else {
		season += fmt.Sprintf("%02d", n.Season)
	}
	return filepath.Join(n.GetSeriesPath(destination), season)
}

// GetMediaPath gives the full filename of given media
func (n EpisodeDetails) GetMediaPath(destination string) string {
	cleanTitle := FileNameCleaner(n.Title)
	cleanShow := FileNameCleaner(n.Showtitle)
	var episode string
	if n.Episode > 0 {
		episode = fmt.Sprintf("s%02de%02d", n.Season, n.Episode)
	} else {
		episode = n.Aired.Time().Format("2006-01-02")
	}

	return filepath.Join(n.GetSeasonPath(destination), cleanShow+" - "+episode+" - "+cleanTitle+".mp4")
}

// GetMediaPathMatcher gives a name matcher for mis numbered episodes
func (n EpisodeDetails) GetMediaPathMatcher(destination string) string {
	cleanTitle := FileNameCleaner(n.Title)
	cleanShow := FileNameCleaner(n.Showtitle)
	return filepath.Join(n.GetSeriesPath(destination), "*", cleanShow+" - * - "+cleanTitle+".mp4")

}

// GetNFOPath give the path where the episode's NFO should be
func (n EpisodeDetails) GetNFOPath(destination string) string {
	nf := n.GetMediaPath(destination)
	return strings.TrimSuffix(nf, filepath.Ext(nf)) + ".nfo"
}

// GetShowNFOPath returns the path for TVShow.nfo
func (n EpisodeDetails) GetShowNFOPath(destination string) string {
	return filepath.Join(destination, FileNameCleaner(n.Showtitle), "tvshow.nfo")
}

// GetShowNFOPath returns the path for TVShow.nfo
func (n EpisodeDetails) GetSeasonNFOPath(destination string) string {
	return filepath.Join(destination, FileNameCleaner(n.Showtitle), fmt.Sprintf("Season %02d", n.Season), "season.nfo")
}

// WriteNFO file at expected place
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
