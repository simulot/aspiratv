package models

import "time"

type Settings struct {
	LastSaved                 time.Time
	LibraryPath               string
	DefaultSeriesSettings     *PathSettings
	DefaultTVShowsSettings    *PathSettings
	DefaultCollectionSettings *PathSettings
	Providers                 map[string]ProvidersSetting
}

type ProvidersSetting struct {
	Name     string
	Disabled bool
	// Account  string
	// Password string // TODO: Secure that!
	Values map[string]string
}

// PathSettings hold templates for naming folder and files of the media
type PathSettings struct {
	ShowPathTemplate      string    // Template for folder that will contain show's files
	SeasonPathTemplate    string    // Template for folder that will contain season's files
	MediaFileNameTemplate string    // Template for the media file name
	FileNamer             FileNamer `json:"-"`
}

// FileNamer implement the code for applying templates
type FileNamer interface {
	ShowPath(info MediaInfo) (string, error)
	SeasonPath(info MediaInfo) (string, error)
	MediaFileName(info MediaInfo) (string, error)
}
