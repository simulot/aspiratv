package frontend

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/simulot/aspiratv/pkg/models"
	"github.com/simulot/aspiratv/pkg/myhttp"
	"github.com/simulot/aspiratv/pkg/providers"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const (
	APIURL      = "/api/"
	providerURL = APIURL + "providers/"
	searchURL   = APIURL + "search/"
)

// API implements a store using RestAPI.
type API struct {
	endPoint string
	client   *myhttp.Client
}

func NewRestStore(endPoint string) *API {
	return &API{
		endPoint: endPoint,
		client: myhttp.NewClient(
			myhttp.WithLogger(log.Default()),
		),
	}
}

func (s *API) ProviderDescribe(ctx context.Context) ([]providers.Description, error) {
	req, err := s.client.NewRequestJSON(ctx, s.endPoint+"providers/", nil, nil, nil)
	if err != nil {
		return nil, err
	}

	var p []providers.Description
	err = s.client.GetJSON(req, &p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *API) Search(ctx context.Context, q models.SearchQuery) (<-chan models.SearchResult, error) {
	// ctx, cancel := context.WithTimeout(ctx, time.Minute)
	ctx, cancel := context.WithCancel(ctx)
	log.Printf("[HTTPCLIENT] Dial websocket %s", s.endPoint+"search/")
	c, _, err := websocket.Dial(ctx, s.endPoint+"search/", nil)
	if err != nil {
		log.Printf("Search Dial error:%s", err)
		cancel()
		return nil, err
	}

	results := make(chan models.SearchResult, 1)

	go func() {
		defer cancel()
		defer close(results)
		defer c.Close(websocket.StatusInternalError, "the sky is falling to the rest client")

		err = wsjson.Write(ctx, c, q)
		if err != nil {
			c.Close(websocket.StatusInternalError, "Search Write Error")
			return
		}

		var status string
		err = wsjson.Read(ctx, c, &status)
		if err != nil {
			c.Close(websocket.StatusInternalError, "Search Read Error")
			return
		}
		if status != "OK" {
			err = fmt.Errorf("Search returns an error: %s", status)
			log.Print(err)
			return
		}

		for {
			select {
			case <-ctx.Done():
				log.Printf("Receive cancellation while writing WS")
				return
			default:
				r := models.SearchResult{}
				if err := wsjson.Read(ctx, c, &r); err != nil {
					var wsErr websocket.CloseError
					if errors.As(err, &wsErr) && wsErr.Code == websocket.StatusNormalClosure {
						c.Close(websocket.StatusNormalClosure, "")
						return
					}
					// TODO log errors
					log.Printf("Can't read WS:%s", err)
					c.Close(websocket.StatusInternalError, "Search read error 2")
					return
				}
				results <- r
			}
		}
	}()

	return results, nil
}

func (s *API) GetSettings(ctx context.Context) (models.Settings, error) {
	req, err := s.client.NewRequestJSON(ctx, s.endPoint+"settings/", nil, nil, nil)
	if err != nil {
		return models.Settings{}, err
	}

	var settings models.Settings

	err = s.client.GetJSON(req, &settings)
	if err != nil {
		return models.Settings{}, err
	}
	return settings, err
}

func (s *API) SetSettings(ctx context.Context, settings models.Settings) (models.Settings, error) {
	req, err := s.client.NewRequestJSON(ctx, s.endPoint+"settings/", nil, nil, settings)
	if err != nil {
		return models.Settings{}, err
	}
	err = s.client.PostJSON(req, &settings)
	if err != nil {
		return models.Settings{}, err
	}
	return settings, err
}

func (s *API) PostDownload(ctx context.Context, dr models.DownloadTask) (models.DownloadTask, error) {
	req, err := s.client.NewRequestJSON(ctx, s.endPoint+"download/", nil, nil, dr)
	if err != nil {
		return models.DownloadTask{}, err
	}
	err = s.client.PostJSON(req, &dr)
	if err != nil {
		return models.DownloadTask{}, err
	}
	return dr, err

}

// Subscribe to server notifications api, return a channel of messages, a closing function and an error
func (s *API) SubscribeServerNotifications(ctx context.Context) (<-chan models.Message, error) {

	ctx, cancel := context.WithCancel(ctx)
	// ctx, cancel := context.WithCancel(ctx)
	log.Printf("[HTTPCLIENT] Dial websocket %s", s.endPoint+"notifications/")
	c, _, err := websocket.Dial(ctx, s.endPoint+"notifications/", nil)
	if err != nil {
		log.Printf("notifications Dial error:%s", err)
		cancel()
		return nil, err
	}
	log.Printf("[HTTPCLIENT] connected to %s", s.endPoint+"notifications/")
	messages := make(chan models.Message, 1)

	go func() {
		defer close(messages)
		defer c.Close(websocket.StatusInternalError, "the sky is falling to the Notification client")
		defer cancel()
		for {
			log.Printf("Notifications WS loop")
			select {
			case <-ctx.Done():
				log.Printf("Receive cancellation while writing WS")
				c.Close(websocket.StatusNormalClosure, "Context cancellation")
				return
			default:
				m := models.Message{}
				log.Printf("Notifications WS loop - wait message")
				if err := wsjson.Read(ctx, c, &m); err != nil {
					var wsErr websocket.CloseError
					if errors.As(err, &wsErr) && wsErr.Code == websocket.StatusNormalClosure {
						c.Close(websocket.StatusNormalClosure, "")
						return
					}
					// TODO log errors
					log.Printf("Can't read WS:%s", err)
					c.Close(websocket.StatusGoingAway, "WS receive error")
					return
				}
				log.Printf("Publish server message")
				messages <- m
				log.Printf("Publish server message done")
			}
		}
	}()

	return messages, nil
}
