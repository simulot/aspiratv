package backend

import (
	"net/http"

	"github.com/simulot/aspiratv/providers"
)

func (s *Server) providersDescribleHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodGet:
		s.decribeProviders(w, r)
		return
	}
	s.sendError(w, APIError{nil, http.StatusMethodNotAllowed, ""})
}

func (s *Server) decribeProviders(w http.ResponseWriter, r *http.Request) {
	d := []providers.ProviderDescription{}
	for _, p := range s.providers {
		d = append(d, p.ProviderDescribe(r.Context()))
	}
	s.writeJsonResponse(w, d, http.StatusOK)
}
