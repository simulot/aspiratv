package backend

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/simulot/aspiratv/pkg/models"
)

func (s *Server) subscriptionsHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodGet:
		s.getSubscriptions(w, r)
	case http.MethodDelete:
		s.deleteSubscription(w, r)
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		s.setSubscriptions(w, r)
	default:
		s.sendError(w, APIError{nil, http.StatusMethodNotAllowed, ""})
	}
}

func (s *Server) deleteSubscription(w http.ResponseWriter, r *http.Request) {
	var UUID uuid.UUID
	var err error

	if p := r.URL.Query().Get("UUID"); p != "" {
		UUID, err = uuid.Parse(p)
		if err != nil {
			s.sendError(w, APIError{err: err, code: http.StatusBadRequest})
			return
		}
	}
	err = s.store.DeleteSubscription(r.Context(), UUID)
	if err != nil {
		s.sendError(w, APIError{err: err, code: http.StatusBadRequest})
		return
	}
	w.Write([]byte("OK"))
}

func (s *Server) getSubscriptions(w http.ResponseWriter, r *http.Request) {
	var UUID uuid.UUID
	var err error

	if p := r.URL.Query().Get("UUID"); p != "" {
		UUID, err = uuid.Parse(p)
		if err != nil {
			s.sendError(w, APIError{err: err, code: http.StatusBadRequest})
			return
		}
	}

	if UUID != uuid.Nil {
		sub, err := s.store.GetSubscription(r.Context(), UUID)
		if err != nil {
			s.sendError(w, err)
		}
		s.writeJsonResponse(w, sub, http.StatusOK)
		return
	}

	subs, err := s.store.GetAllSubscriptions(r.Context())
	if err != nil {
		s.sendError(w, err)
		return
	}
	s.writeJsonResponse(w, subs, http.StatusOK)
}

func (s *Server) setSubscriptions(w http.ResponseWriter, r *http.Request) {
	var (
		sub models.Subscription
		err error
	)

	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(&sub)
	if err != nil {
		s.sendError(w, err)
		return
	}
	sub, err = s.store.SetSubscription(r.Context(), sub)
	if err != nil {
		s.sendError(w, err)
		return
	}
	s.writeJsonResponse(w, sub, http.StatusOK)
}
