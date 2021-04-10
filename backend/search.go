package backend

import (
	"context"
	"log"
	"net/http"

	"github.com/simulot/aspiratv/store"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func (s *APIServer) searchHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Search API called")
	switch r.Method {
	case http.MethodGet:
		err := s.getSearch(w, r)
		if err != nil {
			s.sendError(w, err)
		}
		return
	default:
		s.sendError(w, APIError{nil, http.StatusMethodNotAllowed, ""})
	}
}

func (s *APIServer) getSearch(w http.ResponseWriter, r *http.Request) (err error) {
	ctx, cancel := context.WithCancel(r.Context())

	var results <-chan store.SearchResult
	var c *websocket.Conn
	defer func() {
		cancel()
		if c != nil && err != nil {
			// Prevent sending http error through classical connection when being hijacked
			s.logError(err)
			err = nil
			c.Close(websocket.StatusInternalError, "the sky is falling")
		}
	}()
	results, err = s.store.Search(ctx)
	if err != nil {
		return err
	}
	c, err = websocket.Accept(w, r, nil)
	if err != nil {
		return err
	}

	var query string
	err = wsjson.Read(ctx, c, &query)

	err = s.sendSearchResults(ctx, c, results)
	if err != nil {
		return err
	}
	return c.Close(websocket.StatusNormalClosure, "no more result")
}

func (s *APIServer) sendSearchResults(ctx context.Context, c *websocket.Conn, results <-chan store.SearchResult) error {
	var err error
	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return err

		case r, ok := <-results:
			if !ok {
				return nil
			}
			err = wsjson.Write(ctx, c, r)
			if err != nil {
				// TODO log error
				return err
			}
		}
	}
}
