// htmlparser package wrap colly dependance and let colly using default aspiratv client and its cookies jar

package htmlparser

import (
	"io"

	httplocal "github.com/simulot/aspiratv/net/http"

	"net/http"
	"net/http/cookiejar"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/debug"
)

type Getter interface {
	Get(uri string) (io.Reader, error)
}
type Factory struct {
	jar          *cookiejar.Jar
	roundTripper http.RoundTripper
	userAgent    string
	debugger     debug.Debugger
}

func SetCookieJar(jar *cookiejar.Jar) func(f *Factory) {
	return func(f *Factory) {
		f.jar = jar
	}
}

func SetUserAgent(userAgent string) func(f *Factory) {
	return func(f *Factory) {
		f.userAgent = userAgent
	}
}

func SetTransport(rt http.RoundTripper) func(f *Factory) {
	return func(f *Factory) {
		f.roundTripper = rt
	}
}

func SetDebugger(d debug.Debugger) func(f *Factory) {
	return func(f *Factory) {
		f.debugger = d
	}
}

func NewFactory(conf ...func(f *Factory)) *Factory {
	f := &Factory{
		jar:       httplocal.DefaultClient.Jar,
		userAgent: httplocal.UserAgent,
	}
	for _, fn := range conf {
		fn(f)
	}
	return f
}

func (f *Factory) New() *colly.Collector {
	c := colly.NewCollector()
	if f.debugger != nil {
		c.SetDebugger(f.debugger)
	}
	if len(f.userAgent) > 0 {
		c.UserAgent = f.userAgent
	}
	if f.jar != nil {
		c.SetCookieJar(f.jar)
	}
	if f.roundTripper != nil {
		c.WithTransport(f.roundTripper)
	}
	return c
}
