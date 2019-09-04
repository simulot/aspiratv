package gulli

import (
	"path/filepath"
	"testing"

	"github.com/simulot/aspiratv/parsers/htmlparser"

	"github.com/simulot/aspiratv/net/http/httptest"
)

func Test_Catalog(t *testing.T) {
	p, err := New()
	if err != nil {
		t.Error(err)
	}

	p.htmlParserFactory = htmlparser.NewFactory(htmlparser.SetTransport(
		httptest.New(
			httptest.WithURLToFile(func(u string) string {
				return filepath.Join("testdata", "replay.html.txt")
			}),
		),
	))

	cat, err := p.downloadCatalog()
	if err != nil {
		t.Errorf("Unexpected error %s", err)
		return
	}

	if len(cat) != 49 {
		t.Errorf("Catalog expected to have %d item, but got %d", 49, len(cat))
	}

	_ = cat
}
