package backend

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/simulot/aspiratv/pkg/models"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func (s *Server) notificationsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		err := s.getNotifications(w, r)
		if err != nil {
			s.sendError(w, err)
		}
		return
	default:
		s.sendError(w, APIError{nil, http.StatusMethodNotAllowed, ""})
	}
}

func (s *Server) getNotifications(w http.ResponseWriter, r *http.Request) (err error) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	var c *websocket.Conn
	c, err = websocket.Accept(w, r, nil)
	if err != nil {
		return err
	}
	defer c.Close(websocket.StatusInternalError, "the sky is falling on notification server")

	notificationsChan := make(chan models.Message, 10)
	defer close(notificationsChan)

	cancelSubscription := s.dispatcher.Subscribe(func(p models.Message) {
		notificationsChan <- p
	})
	defer cancelSubscription()

	defer func() {
		log.Printf("server.getNotifications defer")
		if err != nil {
			// Prevent sending http error through classical connection when being hijacked
			s.logError(err)
			err = nil
		}
	}()

	for {
		log.Printf("getNotifications loop")
		select {
		case m := <-notificationsChan:
			log.Printf("getNotifications write message %v", m)
			err = wsjson.Write(ctx, c, m)
			// err = writemessage(ctx, c, m)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			log.Printf("getNotifications context done: %s", ctx.Err())
			return ctx.Err()
		}
	}
}

func writemessage(ctx context.Context, c *websocket.Conn, m models.Message) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return wsjson.Write(ctx, c, m)
}
