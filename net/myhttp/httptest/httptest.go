package httptest

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/simulot/aspiratv/mylog"
)

// HTTPTest provides a HTTP getter to mock http requests
type HTTPTest struct {
	UrlToFilefn func(u string) string
}

// New create a HTTPTest and configures it
func New(conf ...func(ht *HTTPTest)) *HTTPTest {
	ht := &HTTPTest{}
	fileDirect()(ht) // default: url is file name
	for _, fn := range conf {
		fn(ht)
	}
	return ht
}

// the url is the file name
func fileDirect() func(ht *HTTPTest) {
	return func(ht *HTTPTest) {
		ht.UrlToFilefn = func(u string) string {
			return u
		}
	}
}

// WithURLToFile set the custom function UrlToFile
func WithURLToFile(fn func(u string) string) func(ht *HTTPTest) {
	return func(ht *HTTPTest) {
		ht.UrlToFilefn = fn
	}
}

// WithConstantFile read always the same file
func WithConstantFile(s string) func(ht *HTTPTest) {
	return func(ht *HTTPTest) {
		ht.UrlToFilefn = func(string) string {
			return s
		}
	}
}

// RoundTrip implements the file roundtripper and use the UrlToFile function
// to determine the actual file name from the given url
func (ht *HTTPTest) RoundTrip(r *http.Request) (*http.Response, error) {
	url := ""
	if r != nil && r.URL != nil {
		url = r.URL.String()
	}
	f, err := os.Open(ht.UrlToFilefn(url))
	if err != nil {
		return nil, fmt.Errorf("FileTransport.RoundTrip: %v", err)
	}
	fi, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("FileTransport.RoundTrip: %v", err)
	}

	header := make(http.Header)
	header.Add("Content-Type", "html")
	resp := &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.0",
		ProtoMajor:    1,
		ProtoMinor:    0,
		Body:          f,
		ContentLength: fi.Size(),
		Close:         true,
		Request:       r,
		Header:        header,
	}

	return resp, nil
}

// Get implement the Getter interface
func (ht *HTTPTest) Get(u string) (io.ReadCloser, error) {
	url, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	req := &http.Request{
		Method: "GET",
		URL:    url,
	}
	resp, err := ht.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("Can't get url: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Can't get response :%s", resp.Status)
	}

	return resp.Body, nil
}

type myCloser struct {
	io.Reader
	w io.WriteCloser
}

func (mc *myCloser) Close() error {
	if mc.w != nil {
		mc.w.Close()
	}
	if c, ok := mc.Reader.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

func DumpReaderToFile(log *mylog.MyLog, r io.Reader, prefix string) io.ReadCloser {
	f, err := ioutil.TempFile(os.TempDir(), prefix)
	if err != nil {
		log.Fatal().Printf("DumpReaderToFile: %w", err)
	}
	log.Debug().Printf("Reader dumped in %s", f.Name())
	return &myCloser{
		Reader: io.TeeReader(r, f),
		w:      f,
	}
}
