package backend

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/simulot/aspiratv/models"
	"github.com/simulot/aspiratv/providers"
)

func TestSettingsApi(t *testing.T) {
	t.Run("Test GetSettings", func(t *testing.T) {
		spyP := &spyProvider{
			settings: map[string]models.ProvidersSetting{
				"provider": {
					Name: "provider",
				},
			},
		}
		spySt := &spyStore{}
		s := NewServer(spySt, []providers.Provider{spyP})

		request, _ := http.NewRequest(http.MethodGet, "/api/settings/", nil)
		response := httptest.NewRecorder()
		s.ServeHTTP(response, request)

		if response.Code != 200 {
			t.Errorf("Got response code %d, expecting 200", response.Code)
		}

		if !spySt.getSettingsCalled {
			t.Errorf("Expecting GetSettings called")
		}
	})

	t.Run("Test SetSettings", func(t *testing.T) {
		spyP := &spyProvider{
			settings: map[string]models.ProvidersSetting{
				"provider": {
					Name: "provider",
				},
			},
		}
		spySt := &spyStore{}
		s := NewServer(spySt, []providers.Provider{spyP})

		settings := models.Settings{}
		b := bytes.NewBuffer(nil)
		err := json.NewEncoder(b).Encode(&settings)
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
			return
		}

		request, _ := http.NewRequest(http.MethodPost, "/api/settings/", b)
		response := httptest.NewRecorder()
		s.ServeHTTP(response, request)

		if response.Code != 200 {
			t.Errorf("Got response code %d, expecting 200", response.Code)
		}

		if !spySt.setSettingsCalled {
			t.Errorf("Expecting SetSettings called")
		}
	})
}

type spyStore struct {
	getSettingsCalled bool
	setSettingsCalled bool
}

func (s *spyStore) GetSettings() (models.Settings, error) {
	s.getSettingsCalled = true
	return models.Settings{}, nil
}

func (s *spyStore) SetSettings(settings models.Settings) (models.Settings, error) {
	s.setSettingsCalled = true
	return settings, nil
}
