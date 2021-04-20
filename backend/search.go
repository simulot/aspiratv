package backend

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/simulot/aspiratv/models"
	"github.com/simulot/aspiratv/providers"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) getSearch(w http.ResponseWriter, r *http.Request) (err error) {
	ctx, cancel := context.WithCancel(r.Context())

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

	c, err = websocket.Accept(w, r, nil)
	if err != nil {
		return err
	}

	var query models.SearchQuery

	err = wsjson.Read(ctx, c, &query)
	if err != nil {
		log.Printf("Can't decode query: %s", err)
		wsjson.Write(ctx, c, "Error: invalid query")
		return err
	}

	err = wsjson.Write(ctx, c, "OK")
	if err != nil {
		log.Printf("Can't acknowledge query: %s", err)
		return err
	}

	wg := sync.WaitGroup{}

	for _, p := range s.providers {
		wg.Add(1)
		go func(p providers.Provider) {
			defer wg.Done()
			cResults, err := p.Search(ctx, query)
			if err != nil {
				// TODO log error
				return
			}
			for {
				select {
				case <-ctx.Done():
					log.Println("Search cancelled")
					return
				case r, ok := <-cResults:
					if !ok {
						return
					}
					if r.Err != nil {
						log.Printf("Can't send search result: %s", r.Err)
						return
					}

					err = wsjson.Write(ctx, c, r)
					if err != nil {
						log.Printf("Can't send search result: %s", err)
						return
					}

				}
			}

		}(p)

	}
	wg.Wait()
	return c.Close(websocket.StatusNormalClosure, "no more result")
}
