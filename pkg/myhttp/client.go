package myhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

/*
	Define an HTTP Client with suitable defaults and some helpers:
	   - URL construction with path and variables
	   - Easy JSON processing
	   - Rate limiter

*/

type Logger interface {
	Printf(fmt string, a ...interface{})
}

type payloadLogger struct {
	Logger
}

func NewPayloadLogger(logger Logger) *payloadLogger {
	p := payloadLogger{
		Logger: logger,
	}
	return &p
}

func (p *payloadLogger) Printf(f string, a ...interface{}) {
	b, ok := a[len(a)-1].([]byte)
	if !ok {
		p.Logger.Printf(f, append([]interface{}{}, a[:len(a)-1], "-- NO PAYLOAD --")...)
	} else {
		p.Logger.Printf(f, append([]interface{}{}, a[:len(a)-1], string(b))...)
	}
}

type payloadDumper struct {
	d       string
	pattern string
	Logger
	transFn func(b []byte) []byte
}

func NewPayloadDumper(logger Logger, destination string, pattern string, transFn func(b []byte) []byte) *payloadDumper {
	p := payloadDumper{
		Logger:  logger,
		d:       destination,
		pattern: pattern,
		transFn: transFn,
	}
	return &p
}
func (p *payloadDumper) Printf(f string, a ...interface{}) {
	var args []interface{}
	if len(a) > 2 {
		args = a[:len(a)-1]
	}
	b, ok := a[len(a)-1].([]byte)
	if !ok {
		p.Logger.Printf(f, append(args, "-- NO PAYLOAD --")...)
	} else {
		w, err := os.CreateTemp(p.d, p.pattern)
		if err != nil {
			p.Logger.Printf(f, append(args, fmt.Errorf("can't create dump file: %s", err))...)
			return
		}
		defer w.Close()
		if p.transFn != nil {
			b = p.transFn(b)
		}
		w.Write(b)
		p.Logger.Printf(f, append(args, fmt.Sprintf("payload : %d bytes written in file %q", len(b), w.Name()))...)
	}
}

type Client struct {
	http.Client

	logger         Logger
	responseLogger Logger
	requestLogger  Logger
}

type Error struct {
	Err        error  // Original error
	StatusCode int    // HTTP error
	Message    string // Error context
}

func (e Error) Error() string {
	return fmt.Sprintf("%s (%s,%d,%s)", e.Message, e.Err, e.StatusCode, http.StatusText(e.StatusCode))
}

func WithLogger(logger Logger) func(c *Client) {
	return func(c *Client) {
		c.logger = logger
	}
}

func WithRequestLogger(logger Logger) func(c *Client) {
	return func(c *Client) {
		c.requestLogger = logger
	}
}

func WithResponseLogger(logger Logger) func(c *Client) {
	return func(c *Client) {
		c.responseLogger = logger
	}
}

func NewClient(confFn ...func(c *Client)) *Client {
	c := Client{
		logger: log.Default(),
	}
	for _, fn := range confFn {
		fn(&c)
	}

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

func (c *Client) NewRequestJSON(ctx context.Context, URLfmt string, urlParams []interface{}, urlValues url.Values, payload interface{}) (*http.Request, error) {
	var (
		bodyReader io.Reader
		err        error
		b          []byte
	)

	if payload != nil {
		b, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(b)
	}

	if c.requestLogger != nil {
		ctx = context.WithValue(ctx, "request", b)
	}

	req, err := c.NewRequest(ctx, URLfmt, urlParams, urlValues, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-type", "application/json")
	return req, nil
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	// TDODO req.Header.Set("Accept-Encoding", "gzip")

	if c.requestLogger != nil {
		v := req.Context().Value("request")
		if v != nil {
			b, _ := v.([]byte)
			c.requestLogger.Printf("[HTTPCLIENT] %s %s payload: %s", req.Method, req.URL, b)
		}
	} else {
		c.logger.Printf("[HTTPCLIENT] %s %s", req.Method, req.URL)
	}
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

func (c *Client) doJSON(req *http.Request, payload interface{}) error {
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if c.responseLogger != nil {
		c.responseLogger.Printf("[HTTPCLIENT] ... Response: %s(%d), %s", http.StatusText(resp.StatusCode), resp.StatusCode, b)
	}
	return json.Unmarshal(b, payload)
}

func (c *Client) PostJSON(req *http.Request, payload interface{}) error {
	req.Method = http.MethodPost
	return c.doJSON(req, payload)
}

func (c *Client) GetJSON(req *http.Request, payload interface{}) error {
	req.Method = http.MethodGet
	return c.doJSON(req, payload)
}
