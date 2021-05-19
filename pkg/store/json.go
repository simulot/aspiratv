package store

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/simulot/aspiratv/pkg/models"
)

type JSONStore struct {
	l             sync.Mutex                        `json:"lock,omitempty"`
	Settings      models.Settings                   `json:"settings,omitempty"`
	Subscriptions map[uuid.UUID]models.Subscription `json:"subscriptions,omitempty"`
	LastSaved     time.Time                         `json:"last_saved,omitempty"`
	filename      string                            `json:"filename,omitempty"`
}

func NewStoreJSON(filename string) *JSONStore {
	return &JSONStore{
		filename: filename,
	}
}

func freshSettings() models.Settings {
	settings := models.Settings{}
	settings.LibraryPath, _ = os.UserHomeDir()
	settings.LibraryPath = filepath.Join(settings.LibraryPath, "Videos")
	settings.SeriesSettings = models.PathSettings{
		FileNamer:  models.DefaultFileNamer[models.PathTypeSeries],
		Folder:     "Series",
		PathNaming: models.PathTypeSeries,
	}
	settings.CollectionSettings = models.PathSettings{
		FileNamer:  models.DefaultFileNamer[models.PathTypeCollection],
		Folder:     "Collections",
		PathNaming: models.PathTypeCollection,
	}
	settings.TVShowsSettings = models.PathSettings{
		FileNamer:  models.DefaultFileNamer[models.PathTypeTVShow],
		Folder:     "TV",
		PathNaming: models.PathTypeTVShow,
	}
	return settings
}

func (s *JSONStore) GetSettings(ctx context.Context) (models.Settings, error) {
	s.l.Lock()
	defer s.l.Unlock()
	if s.LastSaved.IsZero() {
		err := s.readConfigFile()
		if err != nil {
			return s.Settings, err
		}
	}
	return s.Settings, nil
}

func (s *JSONStore) SetSettings(ctx context.Context, settings models.Settings) (models.Settings, error) {
	s.l.Lock()
	defer s.l.Unlock()

	s.LastSaved = time.Now()
	s.Settings = settings
	err := s.writeConfigFile()
	if err != nil {
		return models.Settings{}, err
	}
	return s.Settings, nil
}

func (s *JSONStore) GetSubscription(ctx context.Context, UUID uuid.UUID) (models.Subscription, error) {
	s.l.Lock()
	defer s.l.Unlock()

	if s.LastSaved.IsZero() {
		err := s.readConfigFile()
		if err != nil {
			return models.Subscription{}, err
		}
	}
	sub, ok := s.Subscriptions[UUID]
	if !ok {
		return models.Subscription{}, ErrorNotFound
	}

	return sub, nil
}
func (s *JSONStore) GetAllSubscriptions(ctx context.Context) ([]models.Subscription, error) {
	s.l.Lock()
	defer s.l.Unlock()

	if s.LastSaved.IsZero() {
		err := s.readConfigFile()
		if err != nil {
			return nil, err
		}
	}
	ss := []models.Subscription{}
	for _, s := range s.Subscriptions {
		ss = append(ss, s)
	}
	return ss, nil
}

func (s *JSONStore) SetSubscription(ctx context.Context, subscription models.Subscription) (models.Subscription, error) {
	s.l.Lock()
	defer s.l.Unlock()

	if subscription.UUID == uuid.Nil {
		subscription.UUID = uuid.New()
	}
	s.Subscriptions[subscription.UUID] = subscription
	s.LastSaved = time.Now()
	err := s.writeConfigFile()
	if err != nil {
		return models.Subscription{}, err
	}
	return subscription, nil
}

func (s *JSONStore) readConfig(r io.Reader) error {
	var config JSONStore
	err := json.NewDecoder(r).Decode(&config)
	if err != nil {
		return err
	}
	s = &config
	return nil
}

func (s *JSONStore) readConfigFile() error {
	f, err := os.Open(s.filename)
	if err != nil {
		if os.IsNotExist(err) {
			s.Settings = freshSettings()
			return nil
		}
		return err
	}
	defer f.Close()
	return s.readConfig(f)
}

func (s *JSONStore) writeConfig(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(s)
}

func (s *JSONStore) writeConfigFile() error {
	f, err := os.Create(s.filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return s.writeConfig(f)
}
