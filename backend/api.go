package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/simulot/aspiratv/store"
)

const (
	APIURL      = "/api/"
	providerURL = APIURL + "providers/"
	searchURL   = APIURL + "search/"
)

type logger interface {
	Logf(f string, args ...interface{})
}

type APIServer struct {
	http.Handler
	store store.Store
	log   logger
}

func NewAPIServer(store store.Store) *APIServer {

	s := &APIServer{
		store: store,
	}
	router := http.NewServeMux()
	router.Handle(providerURL, http.HandlerFunc(s.providersHandler))
	router.Handle(searchURL, http.HandlerFunc(s.searchHandler))
	s.Handler = router
	return s
}

func (s *APIServer) SetLogger(log logger) *APIServer {
	s.log = log
	return s
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
	if e.message != "" {
		return e.message
	}
	return e.err.Error()
}

func (s *APIServer) logError(err error) {
	if s.log != nil {
		s.log.Logf("APIServer:", err)
	}
}

func (s *APIServer) sendError(w http.ResponseWriter, err error) {
	if apiError, ok := err.(*APIError); ok {
		switch apiError.err {
		case store.ErrorNotFound:
			apiError.code = http.StatusNotFound
		case context.Canceled:
			apiError.code = http.StatusServiceUnavailable
			apiError.message = "Cancelled by client"
		case context.DeadlineExceeded:
			apiError.code = http.StatusRequestTimeout
			apiError.message = "Server's timeout exceeded"
		}
		if apiError.code == 0 {
			apiError.code = http.StatusInternalServerError
		}
		if apiError.message == "" {
			apiError.message = http.StatusText(apiError.code)
		}

		s.logError(err)
		http.Error(w, apiError.message, apiError.code)
		return
	}

	s.logError(err)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

}
