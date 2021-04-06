package frontend

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/simulot/aspiratv/store"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const (
	APIURL      = "/api/"
	providerURL = APIURL + "providers/"
	searchURL   = APIURL + "search/"
)

// RestClient implements a store using RestAPI
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

func (s *RestClient) GetProvider(ctx context.Context, name string) (store.Provider, error) {
	req, err := s.newRequest(ctx, s.endPoint+"providers/%s", []string{name}, nil, nil)
	if err != nil {
		return store.Provider{}, err
	}
	p := store.Provider{}
	err = s.do(http.MethodGet, req, &p)
	if err != nil {
		return store.Provider{}, err
	}
	return p, nil
}

func (s *RestClient) SetProvider(ctx context.Context, p store.Provider) (store.Provider, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return store.Provider{}, err
	}

	req, err := s.newRequest(ctx, s.endPoint+"providers/", nil, bytes.NewReader(b), nil)
	if err != nil {
		return store.Provider{}, err
	}

	newP := store.Provider{}
	err = s.do(http.MethodPost, req, &newP)
	if err != nil {
		return store.Provider{}, err
	}
	return newP, nil
}

func (s *RestClient) GetProviderList(ctx context.Context) ([]store.Provider, error) {
	req, err := s.newRequest(ctx, s.endPoint+"providers/", nil, nil, nil)
	if err != nil {
		return nil, err
	}

	list := []store.Provider{}
	err = s.do(http.MethodGet, req, &list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (s *RestClient) Search(ctx context.Context) (<-chan store.SearchResult, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)

	c, _, err := websocket.Dial(ctx, s.endPoint+"search/", nil)
	if err != nil {
		return nil, err
	}

	err = wsjson.Write(ctx, c, "place your search query here")
	if err != nil {
		c.Close(websocket.StatusInternalError, "the sky is falling to the rest client")
		cancel()
		return nil, err
	}
	results := make(chan store.SearchResult, 1)

	go func() {
		defer close(results)
		defer c.Close(websocket.StatusInternalError, "the sky is falling to the rest client")
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				r := store.SearchResult{}
				if err := wsjson.Read(ctx, c, &r); err != nil {
					var wsErr websocket.CloseError
					if errors.As(err, &wsErr) && wsErr.Code == websocket.StatusNormalClosure {
						c.Close(websocket.StatusNormalClosure, "")
						return
					}
					// TODO log errors
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
