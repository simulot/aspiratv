package providers

import (
	"context"
	"encoding/gob"
	"path/filepath"
)

// Provider is the interface for a provider
type Provider interface {
	Name() string                                            // Provider's name
	Shows(context.Context, []*MatchRequest) chan *Show       // List of available shows that match one of MatchRequest
	GetShowStreamURL(context.Context, *Show) (string, error) // Give video stream url ofr a give show
	DebugMode(bool)                                          // Set debug mode
}

// Configurer interface
type Configurer interface {
	SetConfig(map[string]interface{})
}

var providers = map[string]Provider{}

// Register is called by provider's init to register the provider
func Register(p Provider) {
	providers[p.Name()] = p
}

// List of registered providers
func List() map[string]Provider {
	return providers
}

func init() {
	gob.Register([]*Show{})
}

// showFileNames determines show's file name and a matcher for mis numred epidodes
func showFileNames(s *Show) (showName, showMatcher string) {
	var showPath, seasonPath, episodePath, showPathMatcher, seasonPathMatcher, episodePathMatcher string

	if s.Season == "" && s.Episode == "" && s.Show == "" {
		return FileNameCleaner(s.Title) + ".mp4", FileNameCleaner(s.Title) + ".mp4"
	}
	showPath = PathNameCleaner(s.Show)
	showPathMatcher = showPath

	if s.Season == "" {
		seasonPath = "Season " + s.AirDate.Format("2006")
	} else {
		seasonPath = "Season " + Format2Digits(s.Season)
	}
	seasonPathMatcher = "Season *"

	if s.Episode == "" {
		episodePath = FileNameCleaner(s.Show) + " - " + s.AirDate.Format("2006-01-02")
	} else {
		episodePath = FileNameCleaner(s.Show) + " - s" + Format2Digits(s.Season) + "e" + Format2Digits(s.Episode)
	}
	episodePathMatcher = FileNameCleaner(s.Show) + " - *"

	if s.Episode == "" && (s.Title == "" || s.Title == s.Show) {
		episodePath += " - " + s.ID + ".mp4"
		episodePathMatcher += " - " + s.ID + ".mp4"
	} else {
		if s.Title != "" && s.Title != s.Show {
			episodePath += " - " + FileNameCleaner(s.Title) + ".mp4"
			episodePathMatcher += " - " + FileNameCleaner(s.Title) + ".mp4"
		} else {
			episodePath += ".mp4"
			episodePathMatcher += ".mp4"
		}
	}

	return filepath.Join(showPath, seasonPath, episodePath), filepath.Join(showPathMatcher, seasonPathMatcher, episodePathMatcher)
}

// GetShowFileName return a file name with a path that is compatible with PLEX server:
//   ShowName/Season NN/ShowName - sNNeMM - Episode title
//   Show and Episode names are sanitized to avoid problem when saving on the file system
func GetShowFileName(ctx context.Context, s *Show) string {
	r, _ := showFileNames(s)
	return r

}

// GetShowFileNameMatcher return a file pattern of this show
// used for detecting already got episode even when episode or season is different
func GetShowFileNameMatcher(ctx context.Context, s *Show) string {
	_, r := showFileNames(s)
	return r
}
