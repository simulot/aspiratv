package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/simulot/aspiratv/providers"
)

// Config holds settings from configuration file
type Config struct {
	PullInterval textDuration
	Debug        bool                      // Verbose Log output
	Force        bool                      // True to force reload medias
	Destinations map[string]string         // Mapping of destination path
	ConfigFile   string                    // Name of configuration file
	WatchList    []*providers.MatchRequest // Slice of show matchers
	Headless     bool                      // When true, no progression bar
	Providers    map[string]ProviderConfig
}

type ProviderConfig struct {
	Enabled  bool
	Settings map[string]string
}

// Handle Duration as string for JSON configuration
type textDuration time.Duration

func (t textDuration) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Duration(t).String() + `"`), nil
}

func (t *textDuration) UnmarshalJSON(b []byte) error {
	if b[0] == '"' {
		b = b[1 : len(b)-1]
	}
	v, err := time.ParseDuration(string(b))
	if err != nil {
		return err
	}
	*t = textDuration(v)
	return nil
}

// Almost empty configuration for testing purpose
var defaultConfig = &Config{
	PullInterval: textDuration(time.Hour),
	WatchList: []*providers.MatchRequest{
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
	f, err := os.Create("confing.json")
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
func ReadConfig(configFile string) (*Config, error) {
	f, err := os.Open(configFile)
	if err != nil {
		return nil, fmt.Errorf("Can't open configuration file: %v", err)
	}
	defer f.Close()
	conf := &Config{}
	d := json.NewDecoder(f)
	err = d.Decode(conf)
	if err != nil {
		return nil, fmt.Errorf("Can't decode configuration file: %v", err)
	}
	return conf, nil
}

// ReadConfigOrDie create a stub of config.json when it is missing from disk
func ReadConfigOrDie(cli *Config) *Config {
	conf, err := ReadConfig(cli.ConfigFile)
	if err != nil {
		log.Fatalf("Fatal: %v", err)
	}
	// Set flags comming from CLI
	conf.Debug = cli.Debug
	conf.Force = cli.Force
	conf.ConfigFile = cli.ConfigFile
	conf.Headless = cli.Headless
	return conf
}

// Check the configuration or die
func (c *Config) Check() {
	// Expand paths
	for d, p := range c.Destinations {
		c.Destinations[d] = os.ExpandEnv(p)
	}

	for _, m := range c.WatchList {
		m.Pitch = strings.ToLower(m.Pitch)
		m.Show = strings.ToLower(m.Show)
		m.Title = strings.ToLower(m.Title)
		if _, ok := c.Destinations[m.Destination]; !ok {
			log.Fatalf("Destination %q is not defined into section Destination of %q", m.Destination, c.ConfigFile)
		}
	}

}

func (c *Config) IsProviderActive(p string) bool {
	if pc, ok := c.Providers[p]; ok {
		return pc.Enabled
	}
	return false
}
