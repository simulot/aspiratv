package store

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/simulot/aspiratv/pkg/models"
)

type StoreJSON struct {
	settings models.Settings
	filename string
}

func NewStoreJSON(filename string) *StoreJSON {
	return &StoreJSON{
		filename: filename,
	}
}

func (s *StoreJSON) GetSettings() (models.Settings, error) {
	if s.settings.LastSaved.IsZero() {
		err := s.readSettingFile(s.filename)
		if err != nil {
			settings := models.Settings{}
			settings.LibraryPath, _ = os.UserHomeDir()
			filepath.Join(settings.LibraryPath, "Videos", "apsiratv")
			settings.SeriesSettings = &models.PathSettings{
				FileNamer:  models.DefaultFileNamer[models.PathTypeSeries],
				Folder:     "Series",
				PathNaming: models.PathTypeSeries,
			}
			settings.CollectionSettings = &models.PathSettings{
				FileNamer:  models.DefaultFileNamer[models.PathTypeCollection],
				Folder:     "Collections",
				PathNaming: models.PathTypeCollection,
			}
			settings.TVShowsSettings = &models.PathSettings{
				FileNamer:  models.DefaultFileNamer[models.PathTypeTVShow],
				Folder:     "TV",
				PathNaming: models.PathTypeTVShow,
			}
			s.settings = settings
		}
	}
	return s.settings, nil
}

func (s *StoreJSON) SetSettings(settings models.Settings) (models.Settings, error) {
	settings.LastSaved = time.Now()
	s.settings = settings
	err := s.writeSettingsFile(s.filename)
	if err != nil {
		return models.Settings{}, err
	}
	return s.settings, nil
}

func (s *StoreJSON) readSettings(r io.Reader) error {
	var settings models.Settings
	err := json.NewDecoder(r).Decode(&settings)
	if err != nil {
		return err
	}

	s.settings = settings
	return nil
}

func (s *StoreJSON) readSettingFile(name string) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	return s.readSettings(f)
}

func (s *StoreJSON) writeSettings(w io.Writer) error {
	return json.NewEncoder(w).Encode(s.settings)
}

func (s *StoreJSON) writeSettingsFile(name string) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()
	return s.writeSettings(f)
}
