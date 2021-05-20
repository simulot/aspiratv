package models

import (
	"encoding/json"
)

//go:generate enumer -type=PathNamingType -json
type PathNamingType int

const (
	PathTypeUnknown PathNamingType = iota
	PathTypeCollection
	PathTypeSeries
	PathTypeTVShow
	PathTypeMovie
	PathTypeCustom
)

type Settings struct {
	LibraryPath        string                      `json:"library_path,omitempty"`        // Library root path. All folders are given relative to this path
	SeriesSettings     PathSettings                `json:"series_settings,omitempty"`     // Path settings for series
	TVShowsSettings    PathSettings                `json:"tv_shows_settings,omitempty"`   // Path settings for TV Shows
	CollectionSettings PathSettings                `json:"collection_settings,omitempty"` // Path settings fro Collections
	Providers          map[string]ProvidersSetting `json:"providers,omitempty"`           // Settings specifics to providers
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
	Folder                string         `json:"folder,omitempty"`                   // Destination folder for this kind of media relative to library path
	PathNaming            PathNamingType `json:"path_naming,omitempty"`              // One of standard, or custom
	ShowPathTemplate      string         `json:"show_path_template,omitempty"`       // Custom template for folder that will contain show's files
	SeasonPathTemplate    string         `json:"season_path_template,omitempty"`     // Custom template for folder that will contain season's files
	MediaFileNameTemplate string         `json:"media_file_name_template,omitempty"` // Custom template for the media file name
	FileNamer             *FileNamer     `json:"-"`
}

// UnmarshalJSON PathSettings initialise file namer up on PathNaming Filed
func (p *PathSettings) UnmarshalJSON(b []byte) error {
	type pathSettingsjson PathSettings
	s := pathSettingsjson{}
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	*p = PathSettings(s)
	if p.PathNaming != PathTypeCustom {
		p.FileNamer = DefaultFileNamer[p.PathNaming]
	} else {
		p.FileNamer, err = NewFilesNamer(*p, RegularNameCleaner)
		return err
	}
	return nil
}
