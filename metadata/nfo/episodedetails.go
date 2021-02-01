package nfo

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/simulot/aspiratv/providers/matcher"
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

// getSeasonPath give the path for the series' season
func (n *EpisodeDetails) getSeasonPath() string {
	season := "Season "
	if n.Season <= 0 {
		season += "00"
	} else {
		season += fmt.Sprintf("%02d", n.Season)
	}
	return season
}

// GetMediaPath gives the full filename of given media
func (n EpisodeDetails) GetMediaPath(showPath string) string {
	cleanTitle := FileNameCleaner(n.Title)
	cleanShow := FileNameCleaner(n.Showtitle)
	var episode string
	if n.Episode > 0 {
		episode = fmt.Sprintf("s%02de%02d", n.Season, n.Episode)
	} else {
		episode = n.Aired.Time().Format("2006-01-02")
	}
	if cleanTitle == "" {
		return filepath.Join(showPath, n.getSeasonPath(), cleanShow+" - "+episode+".mp4")

	}

	return filepath.Join(showPath, n.getSeasonPath(), cleanShow+" - "+episode+" - "+cleanTitle+".mp4")
}

// Accepted check if ShowTitle or episode Title matches the filter
func (n EpisodeDetails) Accepted(m *matcher.MatchRequest) bool {
	if m.TitleExclude.Regexp != nil {
		if m.TitleExclude.Regexp.MatchString(n.Showtitle) {
			return false
		}
		if m.TitleExclude.Regexp.MatchString(n.Title) {
			return false
		}
	}
	if m.TitleFilter.Regexp != nil {
		if m.TitleFilter.Regexp.MatchString(n.Showtitle) {
			return true
		}
		if m.TitleFilter.Regexp.MatchString(n.Title) {
			return true
		}
		return false
	}
	return true
}

// GetMediaPathMatcher gives a name matcher for mis numbered episodes
func (n EpisodeDetails) GetMediaPathMatcher(showPath string) string {
	cleanTitle := FileNameCleaner(n.Title)
	cleanShow := FileNameCleaner(n.Showtitle)
	return filepath.Join(showPath, "*", cleanShow+" - * - "+cleanTitle+".mp4")

}

// GetNFOPath give the path where the episode's NFO should be
func (n EpisodeDetails) GetNFOPath(showPath string) string {
	nf := n.GetMediaPath(showPath)
	return strings.TrimSuffix(nf, filepath.Ext(nf)) + ".nfo"
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
