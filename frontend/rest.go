package frontend

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/simulot/aspiratv/models"
	"github.com/simulot/aspiratv/myhttp"
	"github.com/simulot/aspiratv/providers"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const (
	APIURL      = "/api/"
	providerURL = APIURL + "providers/"
	searchURL   = APIURL + "search/"
)

// RestClient implements a store using RestAPI.
type RestClient struct {
	endPoint string
	client   *myhttp.Client
}

func NewRestStore(endPoint string) *RestClient {
	return &RestClient{
		endPoint: endPoint,
		client: myhttp.NewClient(
			myhttp.WithLogger(log.Default()),
		),
	}
}

/*
func (s *RestClient) GetProvider(ctx context.Context, name string) (providers.Provider, error) {
	req, err := s.newRequest(ctx, s.endPoint+"providers/%s", []string{name}, nil, nil)
	if err != nil {
		return providers.Provider{}, err
	}
	p := providers.Provider{}
	err = s.do(http.MethodGet, req, &p)
	if err != nil {
		return providers.Provider{}, err
	}
	return p, nil
}

func (s *RestClient) SetProvider(ctx context.Context, p providers.Provider) (providers.Provider, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return providers.Provider{}, err
	}

	req, err := s.newRequest(ctx, s.endPoint+"providers/", nil, bytes.NewReader(b), nil)
	if err != nil {
		return providers.Provider{}, err
	}

	newP := providers.Provider{}
	err = s.do(http.MethodPost, req, &newP)
	if err != nil {
		return providers.Provider{}, err
	}
	return newP, nil
}

func (s *RestClient) GetProviderList(ctx context.Context) ([]providers.Provider, error) {
	req, err := s.newRequest(ctx, s.endPoint+"providers/", nil, nil, nil)
	if err != nil {
		return nil, err
	}

	list := []providers.Provider{}
	err = s.do(http.MethodGet, req, &list)
	if err != nil {
		return nil, err
	}
	return list, nil
}
*/

func (s *RestClient) ProviderDescribe(ctx context.Context) ([]providers.Description, error) {
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

func (s *RestClient) Search(ctx context.Context, q models.SearchQuery) (<-chan models.SearchResult, error) {
	// ctx, cancel := context.WithTimeout(ctx, time.Minute)
	ctx, cancel := context.WithCancel(ctx)
	log.Printf("[HTTPCLIENT] Dial websocket %s", s.endPoint+"search/")
	c, _, err := websocket.Dial(ctx, s.endPoint+"search/", nil)
	if err != nil {
		log.Printf("Search Dial error:%s", err)
		cancel()
		return nil, err
	}

	err = wsjson.Write(ctx, c, q)
	if err != nil {
		c.Close(websocket.StatusInternalError, "the sky is falling to the rest client")
		cancel()
		return nil, err
	}

	var status string
	err = wsjson.Read(ctx, c, &status)
	if err != nil {
		c.Close(websocket.StatusInternalError, "the sky is falling to the rest client")
		cancel()
		return nil, err
	}
	if status != "OK" {
		err = fmt.Errorf("Search returns an error: %s", status)
		log.Print(err)
		cancel()
		return nil, err
	}

	results := make(chan models.SearchResult, 1)

	go func() {
		defer close(results)
		defer c.Close(websocket.StatusInternalError, "the sky is falling to the rest client")
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				log.Printf("Receive cancellation while writing WS")
				return
			default:
				r := models.SearchResult{}
				// log.Print("Another result")
				// _, b, err := c.Read(ctx)
				// log.Print(string(b))
				// if err == nil {
				// 	err = json.Unmarshal(b, &r)
				// }
				// if err != nil {
				// 	var wsErr websocket.CloseError
				// 	if errors.As(err, &wsErr) && wsErr.Code == websocket.StatusNormalClosure {
				// 		c.Close(websocket.StatusNormalClosure, "")
				// 		return
				// 	}
				// 	// TODO log errors
				// 	log.Printf("Can't read WS:%s", err)
				// 	return

				// }

				if err := wsjson.Read(ctx, c, &r); err != nil {
					var wsErr websocket.CloseError
					if errors.As(err, &wsErr) && wsErr.Code == websocket.StatusNormalClosure {
						c.Close(websocket.StatusNormalClosure, "")
						return
					}
					// TODO log errors
					log.Printf("Can't read WS:%s", err)
					return
				}
				results <- r
			}
		}
	}()

	return results, nil
}

// type httpError struct {
// 	StatusCode int
// 	StatusText string
// }

// func (e httpError) Error() string {
// 	return e.StatusText
// }
