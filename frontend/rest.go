package frontend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/simulot/aspiratv/models"
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
	endPoint   string
	httpClient http.Client
}

func NewRestStore(endPoint string) *RestClient {
	return &RestClient{
		endPoint:   endPoint,
		httpClient: http.Client{},
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

func (s *RestClient) ProviderDescribe(ctx context.Context) ([]providers.ProviderDescription, error) {
	req, err := s.newRequest(ctx, s.endPoint+"providers/", nil, nil, nil)
	if err != nil {
		return nil, err
	}

	var p []providers.ProviderDescription
	err = s.get(req, &p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *RestClient) Search(ctx context.Context, q models.SearchQuery) (<-chan models.SearchResult, error) {
	// ctx, cancel := context.WithTimeout(ctx, time.Minute)
	ctx, cancel := context.WithCancel(ctx)

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

// Lets have a good http client (https://www.youtube.com/watch?v=mpcaJux74Qs&t=1717s)

func (s RestClient) newRequest(ctx context.Context, urlFmt string, pathParam []string, reqBody io.Reader, queryParam *url.Values) (*http.Request, error) {
	u := buildEndPoint(urlFmt, pathParam)

	r, err := http.NewRequestWithContext(ctx, "", u, reqBody)
	if err != nil {
		return nil, err
	}

	if queryParam != nil {
		r.URL.RawQuery = queryParam.Encode()
	}

	// r.Header.Set("Content-Type", "application/json")
	return r, nil
}

func (s RestClient) get(r *http.Request, respBody interface{}) error {
	return s.do(http.MethodGet, r, respBody)
}

func (s RestClient) post(r *http.Request, respBody interface{}) error {
	return s.do(http.MethodPost, r, respBody)
}

func (s RestClient) do(method string, r *http.Request, respBody interface{}) error {
	r.Method = method
	r.Header.Set("Accept", "application/json")
	switch method {
	case http.MethodPut, http.MethodPost, http.MethodPatch:
		if r.Header.Get("Content-Type") == "" {
			r.Header.Set("Content-Type", "application/json")
		}
	}

	// TODO Time out/ Dead line

	var resp *http.Response

	resp, err := s.httpClient.Do(r)
	if err != nil {
		// TODO Log
		return err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		// TODO Log
		return err
	}

	if resp.StatusCode >= 400 {
		// Error is just the text sent by server

		// TODO log error
		return httpError{
			StatusCode: resp.StatusCode,
			StatusText: string(b),
		}
	}
	if respBody == nil {
		return nil
	}

	err = json.Unmarshal(b, respBody)
	if err != nil {
		// TODO Log error
		return err
	}
	return nil
}

func buildEndPoint(urlFmt string, pathParam []string) string {
	params := make([]interface{}, len(pathParam))

	for i, pp := range pathParam {
		params[i] = url.PathEscape(pp)
	}
	return fmt.Sprintf(urlFmt, params...)
}

type httpError struct {
	StatusCode int
	StatusText string
}

func (e httpError) Error() string {
	return e.StatusText
}
