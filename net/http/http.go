// The http package provides an HTTP client for the aplication.
// It handles a common cookie jar and same user agent string

package http

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
)

// DefaultClient is the client
var DefaultClient = NewClient()

const UserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Ubuntu Chromium/66.0.3359.181 Chrome/66.0.3359.181 Safari/537.36"

// Client is the classic http client with a cookie jar and a given user agent string
type Client struct {
	*http.Client
	userAgent string
	Jar       *cookiejar.Jar
}

// SetCookieJar is configuration function to provide a cookie jar to the client
func SetCookieJar(cj *cookiejar.Jar) func(c *Client) {
	return func(c *Client) {
		c.Jar = cj
		c.Client.Jar = cj

	}
}

// SetUserAgent is configuration function to give a user agent string to the client
func SetUserAgent(ua string) func(c *Client) {
	return func(c *Client) {
		c.userAgent = ua
	}
}

// NewClient create an HTTP Client and configure it with a set of config functions
func NewClient(conf ...func(c *Client)) *Client {
	c := &Client{
		Client:    &http.Client{},
		userAgent: UserAgent,
	}

	for _, f := range conf {
		f(c)
	}
	return c
}

// Get establish a GET request and return a reader with the response body
func (c *Client) Get(u string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		err = fmt.Errorf("Can't get url: %v", err)
		log.Println(err)
		return nil, err
	}

	req.Header.Set("User-Agent", c.userAgent)
	resp, err := c.Do(req)
	if err != nil {
		err := fmt.Errorf("Can't get: %v", err)
		log.Println(err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Can't get response :%s", resp.Status)
		log.Println(err)
		return nil, err
	}

	return resp.Body, nil
}
