package backend

import (
	"encoding/json"
	"net/http"

	"github.com/simulot/aspiratv/pkg/models"
)

func (s *Server) settingsHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodGet:
		s.getSettings(w, r)
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		s.setSettings(w, r)
	default:
		s.sendError(w, APIError{nil, http.StatusMethodNotAllowed, ""})
	}
}

func (s *Server) getSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := s.store.GetSettings()
	if err != nil {
		s.sendError(w, err)
		return
	}
	s.writeJsonResponse(w, settings, http.StatusOK)
}

func (s *Server) setSettings(w http.ResponseWriter, r *http.Request) {
	var (
		settings models.Settings
		err      error
	)

	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(&settings)
	if err != nil {
		s.sendError(w, err)
		return
	}
	settings, err = s.store.SetSettings(settings)
	if err != nil {
		s.sendError(w, err)
		return
	}
	s.writeJsonResponse(w, settings, http.StatusOK)
}
