package models

import "time"

type Settings struct {
	LastSaved           time.Time
	LibraryPath         string
	DefaultPathSettings PathSettings
	Providers           map[string]ProvidersSetting
}

type ProvidersSetting struct {
	Name     string
	Disabled bool
	// Account  string
	// Password string // TODO: Secure that!
	Values map[string]string
}

type PathSettings struct {
	ShowPath    string
	SeasonPath  string
	EpisodePath string
}
