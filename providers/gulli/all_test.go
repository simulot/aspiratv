package gulli

import (
	"path/filepath"
	"testing"

	"github.com/simulot/aspiratv/net/http/httptest"
	"github.com/simulot/aspiratv/parsers/htmlparser"
)

func Test_getShowList(t *testing.T) {
	getter := httptest.New()
	fact := htmlparser.NewFactory(
		htmlparser.SetTransport(getter),
	).New()
	gss, err := getAllShowList(fact, "file:"+filepath.Join("testdata", "replay.gulli.all.html.txt"))

	if err != nil {
		t.Error(err)
	}
	_ = gss
}
