package providers

import (
	"context"
	"encoding/gob"
	"os"
	"strings"
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

func GetShowPath(s *Show) string {
	if s.Season == "" && s.Episode == "" && s.Show == "" {
		return ""
	}
	return PathNameCleaner(s.Show)
}

func GetSeasonPath(s *Show) string {
	var seasonPath string
	if s.Season == "" && s.Episode == "" && s.Show == "" {
		return ""
	}
	if s.Season == "" {
		seasonPath = "Season " + s.AirDate.Format("2006")
	} else {
		seasonPath = "Season " + Format2Digits(s.Season)
	}
	return seasonPath
}

func GetSeasonMatcher(s *Show) string {
	if s.Season == "" && s.Episode == "" && s.Show == "" {
		return ""
	}
	return "Season *"
}

func GetEpisodeName(s *Show) string {
	episodeName := ""

	if s.Episode == "" {
		episodeName = FileNameCleaner(s.Show) + " - " + s.AirDate.Format("2006-01-02")
	} else {
		episodeName = FileNameCleaner(s.Show) + " - s" + Format2Digits(s.Season) + "e" + Format2Digits(s.Episode)
	}

	if s.Episode == "" && (s.Title == "" || s.Title == s.Show) {
		episodeName += " - " + s.ID + ".mp4"
	} else {
		if s.Title != "" && s.Title != s.Show {
			episodeName += " - " + FileNameCleaner(s.Title) + ".mp4"
		} else {
			episodeName += ".mp4"
		}
	}
	return episodeName
}

func GetEpisodeMatcher(s *Show) string {
	episodeName := ""

	episodeName = FileNameCleaner(s.Show) + " - *"

	if s.Episode == "" && (s.Title == "" || s.Title == s.Show) {
		episodeName += " - " + s.ID + ".mp4"
	} else {
		if s.Title != "" && s.Title != s.Show {
			episodeName += " - " + FileNameCleaner(s.Title) + ".mp4"
		} else {
			episodeName += ".mp4"
		}
	}
	return episodeName

}

func makePath(parts []string) string {
	path := strings.Builder{}
	for _, part := range parts {
		if len(part) > 0 {
			if path.Len() > 0 {
				path.WriteRune(os.PathSeparator)
			}
			path.WriteString(part)
		}
	}
	return path.String()
}

// GetShowFileName return a file name with a path that is compatible with PLEX server:
//   ShowName/Season NN/ShowName - sNNeMM - Episode title
//   Show and Episode names are sanitized to avoid problem when saving on the file system
func GetShowFileName(ctx context.Context, s *Show) string {
	return makePath([]string{GetShowPath(s), GetSeasonPath(s), GetEpisodeName(s)})
}

// GetShowFileNameMatcher return a file pattern of this show
// used for detecting already got episode even when episode or season is different
func GetShowFileNameMatcher(ctx context.Context, s *Show) string {
	return makePath([]string{GetShowPath(s), GetSeasonMatcher(s), GetEpisodeMatcher(s)})
}
