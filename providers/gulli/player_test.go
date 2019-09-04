package gulli

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/simulot/aspiratv/net/http/httptest"
	"github.com/simulot/aspiratv/providers"
)

func Test_getPlayList(t *testing.T) {
	p, err := New()
	if err != nil {
		t.Error(err)
	}

	parser := httptest.New(
		httptest.WithURLToFile(func(u string) string {
			return filepath.Join("testdata", "player.html.txt")
		}),
	)

	p.getter = parser

	shows, err := p.getPlayer("VOD68995029565000")

	if err != nil {
		t.Error(err)
	}
	if len(shows) != 15 {
		t.Errorf("Expecting %d shows, but got %d", 15, len(shows))
		return
	}
	wShow := providers.Show{
		ID:           "VOD68994328878000",
		Title:        "Et la Terre fut...",
		Show:         "Il était une fois... L'Homme",
		Pitch:        "A la fois drôle et bourrée d'informations, la célèbre collection de séries d'animation créée par Albert Barillé réussit à élargir le champ des connaissances des enfants et à leur inculquer le respect des valeurs humaines. Depuis le début de sa création à la fin des années 70, elle a été diffusée dans le monde entier et primée à de multiples reprises. La grande saga de l'humanité, des origines de la vie à nos jours, à travers le quotidien de Monsieur et Madame tout le monde, contée par les inoubliables Maestro, Pierre, Pierrette, leurs enfants et tous leurs compagnons...L'aventure commence il y a presque cinq milliards d'années. Nous approchons du globe terrestre entouré d'une épaisse couche de nuage. Les volcans en éruption, les fleuves de lave, la foudre, les tremblements de terre, la pluie diluvienne forment une vision d'apocalypse. Puis tout semble s'apaiser et un rayon de soleil perce la brume et s'en va frapper la surface de l'eau, là où tout a commencé... Et tandis que tourne l'horloge du temps, Pierre, le robuste Pithécanthrope adulte qui marche des premiers âges de l'humanité va traverser les époques accompagné du sage Maestro, génie distrait par excellence, et devenir un homme d'aujourd'hui. Le voyage le mène, et les enfants à sa suite, d'une école de taille de silex dirigée par Maestro à une île spatiale de l'an 2150. Dans l'intervalle, il nous fera découvrir, entre autre, les rives du Nil au temps de Ramsès II, la Grèce antique de Périclès, le monde de Mahomet, celui des vikings, les voyages de Marco Polo, l'Angleterre d'Elisabeth 1ère, la Russie de Pierre le Grand, la Révolution française, les années folles...",
		Season:       "1",
		Episode:      "1",
		Channel:      "Gulli",
		Duration:     0,
		ShowURL:      "http://replay.gulli.fr/dessins-animes/Il-etait-une-fois-L-Homme/VOD68994328878000",
		StreamURL:    "http://gulli-hls-replay.akamaized.net/68994328878000/68994328878000_ipad.smil/playlist.m3u8",
		ThumbnailURL: "http://resize1-gulli.ladmedia.fr/rcrop/693,389/img/var/storage/imports/replay/images/460121_0.jpg",
		Provider:     "gulli",
		Destination:  "DEST",
	}
	if !reflect.DeepEqual(&wShow, shows[0]) {
		t.Errorf("Expecting show[0] to be \n%#v\nbut got \n%#v", wShow, shows[0])
	}
}
