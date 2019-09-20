package gulli

// import (
// 	"path/filepath"
// 	"testing"

// 	"github.com/simulot/aspiratv/net/myhttp/httptest"
// 	"github.com/simulot/aspiratv/parsers/htmlparser"
// )

// func Test_Episodes(t *testing.T) {
// 	p, err := New()
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	p.htmlParserFactory = htmlparser.NewFactory(htmlparser.SetTransport(
// 		httptest.New(
// 			httptest.WithURLToFile(func(u string) string {
// 				return filepath.Join("testdata", "Oggy-et-les-cafards2.html.txt")
// 			}),
// 		),
// 	))
// 	entry := ShowEntry{
// 		Title: "Oggy et les cafards",
// 		URL:   "https://replay.gulli.fr/dessins-animes/Oggy-et-les-cafards2/VOD68995029565000",
// 	}
// 	ID, err := p.getFirstEpisodeID(entry)
// 	if err != nil {
// 		t.Errorf("Unexpected error %s", err)
// 		return
// 	}

// 	if ID != "VOD68995029565000" {
// 		t.Errorf("Catalog expected to have %s item, but got %s", "VOD68995029565000", ID)
// 	}
// }
