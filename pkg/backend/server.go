package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/simulot/aspiratv/pkg/dispatcher"
	"github.com/simulot/aspiratv/pkg/models"
	"github.com/simulot/aspiratv/pkg/providers"
	"github.com/simulot/aspiratv/pkg/store"
)

const (
	APIURL           = "/api/"
	providerURL      = APIURL + "providers/"
	searchURL        = APIURL + "search/"
	settingsURL      = APIURL + "settings/"
	downloadURL      = APIURL + "download/"
	notificationsURL = APIURL + "notifications/"
	subscriptionsURL = APIURL + "subscriptions/"
)

type logger interface {
	Printf(f string, args ...interface{})
}

type Server struct {
	http.Handler
	store         store.Store
	providers     []providers.Provider
	backgroundCtx context.Context
	log           logger
	dispatcher    *dispatcher.Dispatcher
}

func NewServer(ctx context.Context, store store.Store, p []providers.Provider) *Server {

	s := &Server{
		store:         store,
		providers:     p,
		backgroundCtx: ctx,
		dispatcher:    dispatcher.NewDispatcher(),
		log:           log.Default(),
	}
	s.dispatcher.Subscribe(s.LogNotifications)
	router := http.NewServeMux()
	router.Handle(providerURL, http.HandlerFunc(s.providersDescribleHandler))
	router.Handle(searchURL, http.HandlerFunc(s.searchHandler))
	router.Handle(settingsURL, http.HandlerFunc(s.settingsHandler))
	router.Handle(downloadURL, http.HandlerFunc(s.downloadHandler))
	router.Handle(notificationsURL, http.HandlerFunc(s.notificationsHandler))
	router.Handle(subscriptionsURL, http.HandlerFunc(s.subscriptionsHandler))
	s.Handler = router
	s.Scheduler(ctx)
	return s
}

func (s *Server) Scheduler(ctx context.Context) {
	go func() {
		log.Printf("[SCHEDULER] Started")
		tick := time.NewTicker(10 * time.Second)
		defer tick.Stop()
		for {
			select {
			case t := <-tick.C:
				s.dispatcher.Publish(models.NewMessage(fmt.Sprintf("It's %s, and I'm alive", t.Format("15:04:05"))).SetStatus(models.StatusInfo))
			case <-ctx.Done():
				return
			}
		}
	}()

}

func (s *Server) WithLogger(log logger) *Server {
	s.log = log
	return s
}

func (s *Server) LogNotifications(p *models.Message) {
	s.log.Printf("[SERVER NOTIFICATION] %s", p)
}

func (s *Server) decodeRequest(r *http.Request, body interface{}) error {
	if r.Header.Get("content-type") != "application/json" {
		return &APIError{nil, http.StatusBadRequest, ""}
	}

	err := json.NewDecoder(r.Body).Decode(body)
	if err != nil {
		return &APIError{err, http.StatusBadRequest, ""}
	}
	return nil
}

func (s *Server) writeJsonResponse(w http.ResponseWriter, respBody interface{}, status int) {
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

func (s *Server) logError(err error) {
	if s.log != nil {
		s.log.Printf("APIServer:", err)
	}
}

func (s *Server) sendError(w http.ResponseWriter, err error) {
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

func (s *Server) Provider(name string) providers.Provider {
	for _, p := range s.providers {
		if p.Name() == name {
			return p
		}
	}
	return nil
}
