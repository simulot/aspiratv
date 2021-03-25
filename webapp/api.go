package webapp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/simulot/aspiratv/store"
)

const (
	serverURL   = "/api/"
	providerUrl = serverURL + "providers/"
)

type APIServer struct {
	http.Handler
	store store.Store
}

func NewAPIServer(store store.Store) *APIServer {

	s := &APIServer{
		store: store,
	}
	router := http.NewServeMux()
	router.Handle(providerUrl, http.HandlerFunc(s.providersHandler))
	s.Handler = router
	return s
}

func (s *APIServer) providersHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodGet:
		providerName := strings.TrimPrefix(r.URL.Path, providerUrl)
		switch {
		case providerName != "":
			s.getProvider(w, providerName)
			return
		default:
			s.getProviderList(w)
		}
	case http.MethodPost:
		s.postProvider(w, r)
		return

	}
	s.sendError(w, APIError{nil, http.StatusBadRequest, ""})
}

func (s *APIServer) getProvider(w http.ResponseWriter, providerName string) {
	p, err := s.store.GetProvider(providerName)
	if err != nil {
		s.sendError(w, &APIError{err, 0, ""})
		return
	}
	s.writeJsonResponse(w, p, http.StatusFound)
}

func (s *APIServer) getProviderList(w http.ResponseWriter) {
	p, err := s.store.GetProviderList()
	if err != nil {
		s.sendError(w, &APIError{err, 0, ""})
		return
	}
	s.writeJsonResponse(w, p, http.StatusFound)
}

func (s *APIServer) postProvider(w http.ResponseWriter, r *http.Request) {
	provider := store.Provider{}
	err := s.decodeRequest(r, &provider)
	if err != nil {
		s.sendError(w, err)
		return
	}

	p, err := s.store.SetProvider(provider)
	if err != nil {
		s.sendError(w, err)
		return
	}
	s.writeJsonResponse(w, p, http.StatusAccepted)
}

func (s *APIServer) decodeRequest(r *http.Request, body interface{}) error {
	if r.Header.Get("content-type") != "application/json" {
		return &APIError{nil, http.StatusBadRequest, ""}
	}

	err := json.NewDecoder(r.Body).Decode(body)
	if err != nil {
		return &APIError{err, http.StatusBadRequest, ""}
	}
	return nil
}

func (s *APIServer) writeJsonResponse(w http.ResponseWriter, respBody interface{}, status int) {
	b := bytes.NewBuffer(nil)
	err := json.NewEncoder(b).Encode(respBody)
	if err != nil {
		s.sendError(w, &APIError{err, http.StatusInternalServerError, ""})
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	w.Write(b.Bytes())
}

type APIError struct {
	err     error  // Deep error
	code    int    // Http status code to return
	message string // Public error message
}

func (e APIError) Error() string {
	return e.err.Error()
}

func (s *APIServer) sendError(w http.ResponseWriter, err error) {
	if apiError, ok := err.(*APIError); ok {
		if apiError.err == store.ErrorNotFound {
			apiError.code = http.StatusNotFound
		}
		if apiError.code == 0 {
			apiError.code = http.StatusInternalServerError
		}
		if apiError.message == "" {
			apiError.message = http.StatusText(apiError.code)
		}

		// Todo log error
		http.Error(w, apiError.message, apiError.code)
		return
	}

	// Todo log error
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

}
