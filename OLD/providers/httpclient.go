package providers

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type HTTPClient struct {
	c ProviderConfig
}

func NewHTTPClient(c ProviderConfig) *HTTPClient {
	return &HTTPClient{
		c: c,
	}
}

func (c *HTTPClient) newRequest(ctx context.Context, url string, body io.Reader, q *url.Values) (*http.Request, error) {
	r, err := http.NewRequestWithContext(ctx, "", url, body)
	if err != nil {
		return nil, err
	}
	if q != nil {
		r.URL.RawQuery = q.Encode()
	}

	if c.c.UserAgent != "" {
		r.Header.Set("User-Agent", c.c.UserAgent)
	}
	return r, nil
}

func (c *HTTPClient) do(ctx context.Context, method string, r *http.Request) ([]byte, error) {
	httpC := &http.Client{
		Timeout: 10 * time.Second,
	}

	if c.c.Log.IsDebug() {
		httpC.Timeout = 10 * time.Minute
	}

	if c.c.HitsLimiter != nil {
		err := c.c.HitsLimiter.Wait(ctx)
		if err != nil {
			return nil, err
		}
	}

	r.Method = method
	resp, err := httpC.Do(r)
	if err != nil {
		c.c.Log.Error().Printf("[%s] %s: %s", r.Method, r.URL.String(), err)
		return nil, err
	}

	defer resp.Body.Close()
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		c.c.Log.Error().Printf("[%s] %s: (%s): %s", r.Method, r.URL.String(), resp.Status, err)
		return nil, err
	}

	if resp.StatusCode >= 400 {
		c.c.Log.Error().Printf("[%s] %s: (%s): %s", r.Method, r.URL.String(), resp.Status, string(buf))
		return nil, err

	}

	if c.c.Log.IsDebug() {
		c.c.Log.Trace().Printf("[%s] %s:%s Content-Type:%q", r.Method, r.URL.String(), resp.Status, resp.Header.Get("content-type"))
		if strings.Contains(resp.Header.Get("content-type"), "json") {
			c.c.Log.Trace().Printf("[%s]\n%s", r.Method, string(buf))
		} else {
			f, err := os.CreateTemp("", "*.dump")
			if err != nil {
				c.c.Log.Fatal().Printf("Can't dump response to file: %w", err)
				return nil, err
			}
			defer f.Close()
			c.c.Log.Debug().Printf("Response dumped in %q", f.Name())
			_, err = f.Write(buf)
		}
	}
	return buf, nil
}

func (c *HTTPClient) Get(ctx context.Context, url string, q *url.Values, body io.Reader) ([]byte, error) {
	r, err := c.newRequest(ctx, url, body, q)
	if err != nil {
		return nil, err
	}
	return c.do(ctx, http.MethodGet, r)
}

func (c *HTTPClient) Head(ctx context.Context, url string, q *url.Values, body io.Reader) ([]byte, error) {
	r, err := c.newRequest(ctx, url, body, q)
	if err != nil {
		return nil, err
	}
	return c.do(ctx, http.MethodHead, r)
}

func (c *HTTPClient) Post(ctx context.Context, url string, q *url.Values, body io.Reader) ([]byte, error) {
	r, err := c.newRequest(ctx, url, body, q)
	if err != nil {
		return nil, err
	}
	return c.do(ctx, http.MethodPost, r)
}
