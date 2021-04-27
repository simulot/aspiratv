package store

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"time"

	"github.com/simulot/aspiratv/models"
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
			return models.Settings{}, err
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
