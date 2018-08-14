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
	Debug        bool                      // Log ffmep output
	Force        bool                      // True to force reload medias
	Service      bool                      // True when runing as service. When false, query all provider en terminate
	Destinations map[string]string         // Mapping of destination path
	WatchList    []*providers.MatchRequest // Slice of show matcher
}

// Handle Duration as string for JSON configation
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

// WriteConfig create a JSON file with the current confuration
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
func ReadConfig() (*Config, error) {
	f, err := os.Open("config.json")
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

func ReadConfigOrGenerateDefault() *Config {
	conf, err := ReadConfig()
	if err != nil {
		log.Printf("%v, generate a default file.", err)
		conf = defaultConfig
		WriteConfig()
	}
	return conf
}

// Check the configuration or die
func (c *Config) Check() {
	for _, m := range c.WatchList {
		m.Pitch = strings.ToLower(m.Pitch)
		m.Show = strings.ToLower(m.Show)
		m.Title = strings.ToLower(m.Title)
	}

	// Expand paths
	for d, p := range c.Destinations {
		c.Destinations[d] = os.ExpandEnv(p)
	}

}
