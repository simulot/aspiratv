package providers

import (
	"strings"
)

type MatchRequest struct {
	// Fields for matching
	Show     string
	Title    string
	Pitch    string
	Provider string

	// Destination name when found
	Destination string
}

func Match(m *MatchRequest, s *Show) bool {
	if len(m.Provider) > 0 {
		if m.Provider != strings.ToLower(s.Provider) {
			return false
		}
	}
	if len(m.Show) > 0 {
		if f := strings.Contains(strings.ToLower(s.Show), m.Show); !f {
			return false
		}
	}
	if len(m.Title) > 0 {
		if f := strings.Contains(strings.ToLower(s.Title), m.Title); !f {
			return false
		}
	}
	if len(m.Pitch) > 0 {
		if f := strings.Contains(strings.ToLower(s.Pitch), m.Pitch); !f {
			return false
		}
	}

	return true
}
