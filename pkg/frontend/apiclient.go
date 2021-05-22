package frontend

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/simulot/aspiratv/pkg/models"
	"github.com/simulot/aspiratv/pkg/myhttp"
	"github.com/simulot/aspiratv/pkg/providers"
	"github.com/simulot/aspiratv/pkg/store"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const (
	settingsURL      = "settings/"
	providerURL      = "providers/"
	searchURL        = "search/"
	downladURL       = "download/"
	notificationsURL = "notifications/"
)

// APIClient implements a store using RestAPI.
type APIClient struct {
	endPoint string
	client   *myhttp.Client
	Store    store.Store
}

func NewAPIClient(endPoint string, s store.Store) *APIClient {
	return &APIClient{
		endPoint: endPoint,
		Store:    s,
		client: myhttp.NewClient(
			myhttp.WithLogger(log.Default()),
		),
	}
}

func (s *APIClient) ProviderDescribe(ctx context.Context) ([]providers.Description, error) {
	req, err := s.client.NewRequestJSON(ctx, s.endPoint+providerURL, nil, nil, nil)
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

func (s *APIClient) Search(ctx context.Context, q models.SearchQuery) (<-chan models.SearchResult, error) {
	// ctx, cancel := context.WithTimeout(ctx, time.Minute)
	ctx, cancel := context.WithCancel(ctx)
	log.Printf("[HTTPCLIENT] Dial websocket %s", s.endPoint+searchURL)
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

func (s *APIClient) GetSettings(ctx context.Context) (models.Settings, error) {
	return s.Store.GetSettings(ctx)
}

func (s *APIClient) SetSettings(ctx context.Context, settings models.Settings) (models.Settings, error) {
	return s.Store.SetSettings(ctx, settings)
}

func (s *APIClient) PostDownload(ctx context.Context, dr models.DownloadTask) (models.DownloadTask, error) {
	req, err := s.client.NewRequestJSON(ctx, s.endPoint+downladURL, nil, nil, dr)
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
func (s *APIClient) SubscribeServerNotifications(ctx context.Context) (<-chan *models.Message, error) {
	ctx, cancel := context.WithCancel(ctx)
	c, _, err := websocket.Dial(ctx, s.endPoint+notificationsURL, nil)
	if err != nil {
		log.Printf("[HTTPCLIENT] Can't connect to %s: %s", s.endPoint+notificationsURL, err)
		cancel()
		return nil, err
	}
	log.Printf("[HTTPCLIENT] connected to %s", s.endPoint+notificationsURL)
	messages := make(chan *models.Message, 1)

	go func() {
		defer close(messages)
		defer c.Close(websocket.StatusInternalError, "the sky is falling to the Notification client")
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				c.Close(websocket.StatusNormalClosure, "Context cancellation")
				return
			default:
				m := models.Message{}
				if err := wsjson.Read(ctx, c, &m); err != nil {
					var wsErr websocket.CloseError
					if errors.As(err, &wsErr) && wsErr.Code == websocket.StatusNormalClosure {
						c.Close(websocket.StatusNormalClosure, "")
						return
					}
					log.Printf("[HTTPCLIENT] Can't read message:%s", err)
					c.Close(websocket.StatusGoingAway, "WS receive error")
					return
				}
				messages <- &m
			}
		}
	}()

	return messages, nil
}

func (s *APIClient) GetSubscription(ctx context.Context, UUID uuid.UUID) (models.Subscription, error) {
	return s.Store.GetSubscription(ctx, UUID)
}

func (s *APIClient) GetAllSubscriptions(ctx context.Context) ([]models.Subscription, error) {
	return s.Store.GetAllSubscriptions(ctx)
}

func (s *APIClient) SetSubscription(ctx context.Context, sub models.Subscription) (models.Subscription, error) {
	return s.Store.SetSubscription(ctx, sub)
}
