package store

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"time"

	"github.com/simulot/aspiratv/pkg/library"
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
			if !errors.Is(err, os.ErrNotExist) {
				log.Printf("[STORE] Can't read %q: %s", s.filename, err)
			} else {
				err = nil
			}
			settings := models.Settings{}
			if err != nil {
				return settings, err
			}
			settings.DefaultSeriesSettings, err = initPathSettings(library.DefaultPathSettings[models.TypeSeries], library.RegularNameCleaner)
			if err != nil {
				return settings, err
			}
			settings.DefaultCollectionSettings, err = initPathSettings(library.DefaultPathSettings[models.TypeCollection], library.RegularNameCleaner)
			if err != nil {
				return settings, err
			}
			settings.DefaultTVShowsSettings, err = initPathSettings(library.DefaultPathSettings[models.TypeTVShow], library.RegularNameCleaner)
			if err != nil {
				return settings, err
			}
			return settings, nil
		}

	}

	return s.settings, nil
}

func initPathSettings(s models.PathSettings, nameCleaner *library.NameCleaner) (*models.PathSettings, error) {
	var err error
	settings := s
	settings.FileNamer, err = library.NewFilesNamer(s, nameCleaner)
	return &settings, err
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
	settings.DefaultSeriesSettings, err = initPathSettings(*settings.DefaultSeriesSettings, library.RegularNameCleaner)
	if err != nil {
		return err
	}
	settings.DefaultCollectionSettings, err = initPathSettings(*settings.DefaultCollectionSettings, library.RegularNameCleaner)
	if err != nil {
		return err
	}
	settings.DefaultTVShowsSettings, err = initPathSettings(*settings.DefaultTVShowsSettings, library.RegularNameCleaner)
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
