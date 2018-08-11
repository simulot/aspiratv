package httptest

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

type HttpTest struct {
	UrlToFilefn func(u string) string
}

func New(conf ...func(ht *HttpTest)) *HttpTest {
	ht := &HttpTest{}
	fileDirect()(ht)
	for _, fn := range conf {
		fn(ht)
	}
	return ht
}

func fileDirect() func(ht *HttpTest) {
	return func(ht *HttpTest) {
		ht.UrlToFilefn = func(u string) string {
			return u
		}
	}
}

func (ht *HttpTest) RoundTrip(r *http.Request) (*http.Response, error) {
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

func (ht *HttpTest) Get(u string) (io.Reader, error) {
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
	// the pipe allows to close the response body and close the pipe
	pr, pw := io.Pipe()
	go func() {
		defer resp.Body.Close()
		_, err := io.Copy(pw, resp.Body)
		if err != nil {
			pw.CloseWithError(fmt.Errorf("Can't get: %v", err))
			return
		}
		pw.Close()
	}()

	return pr, nil
}
