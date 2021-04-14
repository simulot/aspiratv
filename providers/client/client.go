package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

/*
	Define an HTTP Client with suitable defaults and some helpers:
	   - URL construction with path and variables
	   - Easy JSON processing
	   - Rate limiter

*/

type Logger interface {
	Logf(fmt string, args []interface{})
}

type Client struct {
	http.Client
}

type Error struct {
	Err        error  // Original error
	StatusCode int    // HTTP error
	Message    string // Error context
}

func (e Error) Error() string {
	return fmt.Sprintf("%s (%s,%d,%s)", e.Message, e.Err, e.StatusCode, http.StatusText(e.StatusCode))
}

func New() *Client {
	c := Client{}

	return &c
}

func (c *Client) NewRequest(ctx context.Context, URLfmt string, urlParams []interface{}, urlValues url.Values, body io.Reader) (*http.Request, error) {
	u, err := url.Parse(fmt.Sprintf(URLfmt, urlParams...))
	if err != nil {
		return nil, err
	}
	u.RawQuery = urlValues.Encode()
	return http.NewRequestWithContext(ctx, "", u.String(), body)
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		err = Error{
			StatusCode: resp.StatusCode,
			Message:    string(b),
		}
		resp.Body.Close()
		return nil, err
	}
	return resp, nil
}

func (c *Client) Get(req *http.Request) (*http.Response, error) {
	req.Method = http.MethodGet
	return c.Do(req)
}

func (c *Client) GetJSON(req *http.Request, payload interface{}) error {
	// TDODO req.Header.Set("Accept-Encoding", "gzip")
	req.Method = http.MethodGet
	log.Print(req.URL.String())
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, payload)
}
