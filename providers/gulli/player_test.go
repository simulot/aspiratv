package gulli

import (
	"path/filepath"
	"testing"

	"github.com/simulot/aspiratv/net/http/httptest"
)

func Test_getPlayList(t *testing.T) {
	p, err := New()
	if err != nil {
		t.Error(err)
	}

	parser := httptest.New(
		httptest.WithURLToFile(func(u string) string {
			return filepath.Join("testdata", "player_VOD68993752052000.html.txt")
		}),
	)

	p.getter = parser

	shows, err := p.getPlayer("http://replay.gulli.fr/dessins-animes/Oggy-et-les-cafards/VOD68993752052000", "VOD68993752052000")

	if err != nil {
		t.Error(err)
	}
	if len(shows) != 5 {
		t.Errorf("Expecting %d shows, but got %d", 5, len(shows))
		return
	}
	if shows[0].ID != "VOD68993751872000" {
		t.Errorf("Expecting ID for shows 0 to be %s, but got %s", "VOD68993751872000", shows[0].ID)
	}
}
