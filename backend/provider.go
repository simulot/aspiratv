package backend

import (
	"context"
	"net/http"
	"strings"

	"github.com/simulot/aspiratv/store"
)

func (s *APIServer) providersHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodGet:
		providerName := strings.TrimPrefix(r.URL.Path, providerURL)
		switch {
		case providerName != "":
			s.getProvider(r.Context(), w, providerName)
			return
		default:
			s.getProviderList(r.Context(), w)
			return
		}
	case http.MethodPost:
		s.postProvider(r.Context(), w, r)
		return

	}
	s.sendError(w, APIError{nil, http.StatusMethodNotAllowed, ""})
}

func (s *APIServer) getProvider(ctx context.Context, w http.ResponseWriter, providerName string) {
	p, err := s.store.GetProvider(ctx, providerName)
	if err != nil {
		s.sendError(w, &APIError{err, 0, "provider '" + providerName + "' not found"})
		return
	}
	s.writeJsonResponse(w, p, http.StatusOK)
}

func (s *APIServer) getProviderList(ctx context.Context, w http.ResponseWriter) {
	p, err := s.store.GetProviderList(ctx)
	if err != nil {
		s.sendError(w, &APIError{err, 0, ""})
		return
	}
	s.writeJsonResponse(w, p, http.StatusOK)
}

func (s *APIServer) postProvider(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	provider := store.Provider{}
	err := s.decodeRequest(r, &provider)
	if err != nil {
		s.sendError(w, err)
		return
	}

	p, err := s.store.SetProvider(ctx, provider)
	if err != nil {
		s.sendError(w, err)
		return
	}
	s.writeJsonResponse(w, p, http.StatusAccepted)
}
