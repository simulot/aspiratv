package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/simulot/aspiratv/matcher"
	"github.com/simulot/aspiratv/providers"
)

func (a *app) Initialize(cmd string) {

	// Check ffmpeg presence
	var c *exec.Cmd
	if runtime.GOOS == "windows" {
		c = exec.Command("where", "ffmpeg")
	} else {
		c = exec.Command("which", "ffmpeg")
	}
	b, err := c.Output()
	if err != nil {
		a.logger.Fatal().Printf("[Initialize] Can't determine ffmpeg path: %s", err)
	}
	a.ffmpeg = strings.Trim(strings.Trim(string(b), "\r\n"), "\n")
	a.logger.Trace().Printf("[Initialize] FFMPEG path: %q", a.ffmpeg)

	// Get FFMPEG version
	c = exec.Command(a.ffmpeg, "-version")
	b, err = c.Output()
	a.logger.Debug().Printf("[Initialize] FFMPEG version: %q", string(b))
}

// ProviderConfig hold configuration for providers in config.json file
type ProviderConfig struct {
	Enabled  bool
	Settings map[string]string
}

// Almost empty configuration for testing purpose
var defaultConfig = &providers.Settings{
	WatchList: []*matcher.MatchRequest{
		{
			Provider:    "francetv",
			Show:        "Les Lapins Crétins",
			Destination: "Jeunesse",
		},
	},
	Destinations: map[string]string{
		"Jeunesse": "${HOME}/Videos/Jeunesse",
	},
}

// WriteConfig create a JSON file with the current configuration
func WriteConfig() {
	f, err := os.Create("config.json")
	if err != nil {
		log.Fatalf("Can't write configuration file: %v", err)
	}
	e := json.NewEncoder(f)
	e.SetIndent("", "  ")
	e.Encode(defaultConfig)
	f.Close()
	os.Exit(0)
}

// ReadConfig read the JSON configuration file
func (a *app) ReadConfig(configFile string) error {
	a.logger.Trace().Printf("[ReadConfig] opening '%s'", configFile)
	f, err := os.Open(configFile)
	if err != nil {
		return fmt.Errorf("Can't open configuration file: %v", err)
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&a.Settings)
	if err != nil {
		return fmt.Errorf("Can't decode configuration file: %v", err)
	}
	err = a.Settings.CheckPath()
	if err != nil {
		return err
	}
	return nil
}
