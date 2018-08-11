// The http package provides an HTTP client for the aplication.
// It handles a common cookie jar and same user agent string

package http

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func makeJar() *cookiejar.Jar {
	c, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	u, err := url.Parse("http://root.com")
	if err != nil {
		panic(err)
	}
	c.SetCookies(u, []*http.Cookie{&http.Cookie{Name: "Flavor", Value: "Chocolate Chip"}})
	return c
}

func TestNewClient(t *testing.T) {
	type args struct {
		conf []func(c *Client)
	}

	cj := makeJar()

	tests := []struct {
		name string
		args args
		want *Client
	}{
		{
			"default",
			args{},
			&Client{
				&http.Client{},
				UserAgent,
				nil,
			},
		},
		{
			"with agent",
			args{
				[]func(c *Client){SetUserAgent("Given Agent")},
			},
			&Client{
				&http.Client{},
				"Given Agent",
				nil,
			},
		},
		{
			"with cookiejar",
			args{
				[]func(c *Client){SetCookieJar(cj)},
			},
			&Client{
				&http.Client{Jar: cj},
				UserAgent,
				makeJar(),
			},
		}, {
			"with cookiejar and user agent",
			args{
				[]func(c *Client){SetCookieJar(cj), SetUserAgent("Given Agent")},
			},
			&Client{
				&http.Client{Jar: cj},
				"Given Agent",
				makeJar(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.args.conf...)
			if client == nil {
				t.Error("Unexptected nil")
				return
			}
			if client.userAgent != tt.want.userAgent {
				t.Errorf("Want userAgent to be %q, but got %q", tt.want.userAgent, client.userAgent)
			}
			if client.Jar == nil && tt.want.Jar != nil {
				t.Errorf("Want cookie jar to be %v, but got %v", tt.want.Jar, client.Jar)
				return
			}
			if client.Jar != nil && tt.want.Jar == nil {
				t.Errorf("Want cookie jar to be %v, but got %v", tt.want.Jar, client.Jar)
				return
			}
			if client.Jar == nil && tt.want.Jar == nil {
				return
			}
			u, _ := url.Parse("http://root.com")
			if !reflect.DeepEqual(client.Jar.Cookies(u), tt.want.Jar.Cookies(u)) {
				t.Errorf("Want cookie jar %v, but got %v", tt.want.Jar, client.Jar)
			}
		})
	}
}

type tstHandler struct {
	body   *strings.Builder
	status int
}

func newTH(l int, status int) *tstHandler {
	s := &tstHandler{
		status: status,
		body:   &strings.Builder{},
	}
	for i := 0; i < l; i++ {
		s.body.WriteByte(byte(i%26) + 'A')
	}
	return s
}

func (th *tstHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, err := io.Copy(w, strings.NewReader(th.body.String()))
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(th.status)
}

func TestGet(t *testing.T) {
	th := newTH(16*1024*1024, 200)
	ts := httptest.NewServer(th)
	defer ts.Close()

	c := NewClient()
	res, err := c.Get(ts.URL)
	if err != nil {
		t.Errorf("Can't get: %v", err)
		return
	}

	b := &strings.Builder{}
	n, err := io.Copy(b, res)
	if err != nil {
		t.Errorf("Can't read response: %v", err)
		return
	}
	if n != int64(th.body.Len()) {
		t.Errorf("Expected recieve %d bytes, but got %d", th.body.Len(), n)
		return
	}
	if b.String() != th.body.String() {
		t.Errorf("Recieved content differes from expected")
		return
	}
}

func TestClient_Get(t *testing.T) {
	type fields struct {
		Client    *http.Client
		userAgent string
	}
	type args struct {
		u string
	}
	tests := []struct {
		name    string
		testSrv *tstHandler
		wantErr bool
	}{
		{
			"normal",
			newTH(16*1024*1024, 200),
			false,
		},
		{
			"error",
			newTH(0, 404),
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(tt.testSrv)
			c := NewClient()

			got, err := c.Get(ts.URL)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				b := &strings.Builder{}
				_, err = io.Copy(b, got)
				if b.String() != tt.testSrv.body.String() {
					t.Errorf("Recieved content differs from expected")
				}
			}
		})
	}
}
