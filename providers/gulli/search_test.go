package gulli

import (
	"testing"
)

func Test_searchPlayer(t *testing.T) {
	// p, _ := New(
	// 	withGetter(
	// 		httptest.New(httptest.WithURLToFile(func(u string) string {
	// 			return filepath.Join("testdata", "recherche.html.txt")
	// 		}))),
	// )

	// mm := []*providers.MatchRequest{
	// 	{
	// 		Show:        strings.ToLower("Il était une fois... L'Homme"),
	// 		Provider:    "gulli",
	// 		Destination: "dest1",
	// 	},
	// }

	// ID, ShowURL, err := p.searchPlayer(mm[0])

	// if err != nil {
	// 	t.Error(err)
	// }

	// wID := "VOD69001494489000"
	// if ID != wID {
	// 	t.Errorf("Expecting ID to be %s, but got %s", wID, ID)
	// }

	// wURL := "http://replay.gulli.fr/dessins-animes/Il-etait-une-fois-L-Homme/VOD69001494489000"
	// if ShowURL != wURL {
	// 	t.Errorf("Expecting shoURL to be %s, but got %s", wURL, ShowURL)
	// }
}
