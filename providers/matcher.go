package providers

import (
	"strings"
)

// MatchRequest holds criterions for selecting show
type MatchRequest struct {
	// Fields for matching
	Show     string
	ShowID   string
	Title    string
	TitleID  string
	Pitch    string
	Provider string
	Playlist string // Playlist search is implemented in providers.

	// Destination name when found
	Destination string
}

// IsShowMatch is the generic implementation of show matcher.
// Criterions are tested in following order:
// - Provider
// - Show
// - Title
// - Pitch
// When there is a match, it adds  MatchRequest.Destination into Show record.
// Criteria is ignored when it is empty in the MatchRequest
// When the list of MatchRequest is nil or empty, all show will match.
// Note: Playlist match isn't handled generically, it must be implemented in the provider's implementation
func IsShowMatch(mm []*MatchRequest, s *Show) bool {
	if mm == nil || len(mm) == 0 {
		return true
	}
	for _, m := range mm {
		if len(m.Provider) > 0 {
			if m.Provider != strings.ToLower(s.Provider) {
				continue
			}
		}
		if len(m.Show) > 0 {
			if f := strings.Contains(strings.ToLower(s.Show), m.Show); !f {
				continue
			}
		}
		if len(m.Title) > 0 {
			if f := strings.Contains(strings.ToLower(s.Title), m.Title); !f {
				continue
			}
		}
		if len(m.Pitch) > 0 {
			if f := strings.Contains(strings.ToLower(s.Pitch), m.Pitch); !f {
				continue
			}
		}
		s.Destination = m.Destination
		return true
	}
	return false
}
