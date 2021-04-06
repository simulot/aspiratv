package providers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/simulot/aspiratv/matcher"
)

// Settings hold application global settings
type Settings struct {
	Providers    map[string]ProviderSettings // Registered providers
	Destinations map[string]string           // Mapping of destination path
	WatchList    []*matcher.MatchRequest     // Slice of show matchers
	// TODO restore WriteNFO option
	// WriteNFO     bool                        // True when NFO files to be written
}

type ProviderSettings struct { // TODO don't stutter!
	Enabled  bool
	HitsRate int // Number of get per second
	Settings map[string]string
}

func (s *Settings) CheckPath() error {
	for k, v := range s.Destinations {
		var err error

		v, err := ExpandPath(v)
		if err != nil {
			return fmt.Errorf("Can't create destination directory for %q: %w", k, err)
		}

		err = os.MkdirAll(v, 0755)
		if err != nil {
			return fmt.Errorf("Can't create destination directory for %q: %w", k, err)
		}
		s.Destinations[k] = v
	}

	for _, m := range s.WatchList {
		if _, ok := List()[m.Provider]; !ok {
			return fmt.Errorf("Unknown provider %q for show %q", m.Provider, m.Show)
		}
		err := m.Validate(s.Destinations)
		if err != nil {
			return err
		}
		if m.ShowRootPath != "" {
			m.ShowRootPath, err = ExpandPath(m.ShowRootPath)
			if err != nil {
				return err
			}
		}
		m.Show = strings.ToLower(m.Show)
		m.Title = strings.ToLower(m.Title)
	}
	return nil
}

// ExpandPath variables contained in the path
//  - ENV variables like $HOME
//	- ~ for home dir
func ExpandPath(p string) (string, error) {
	if strings.HasPrefix(p, "~/") {
		u, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		p = filepath.Join(u, p[2:])
	}
	p = os.ExpandEnv(p)
	return filepath.Abs(p)
}
