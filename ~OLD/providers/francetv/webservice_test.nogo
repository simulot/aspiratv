package francetv

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStructPluzzEmission(t *testing.T) {
	const js = `{
		"nb_vues": "56",
		"volonte_replay": "1",
		"replay": "1",
		"mandat_duree": "7",
		"etranger": "0",
		"url": "\/videos\/inspecteur_gadget_france_4_saison1_,184533015.html",
		"soustitrage": "1",
		"recurrent": "1",
		"url_video_sitemap": "",
		"temps_restant": "424850",
		"duree_reelle": "605",
		"nb_diffusion": "7",
		"oas_sitepage": "france4\/jeunesse\/inspecteurgadgetfrance4",
		"realisateurs": "",
		"image_200": "\/image\/referentiel_emissions\/184533015\/1533577896\/200\/0\/france4\/0\/img.jpg",
		"image_medium": "\/image\/referentiel_emissions\/184533015\/1533577896\/214\/0\/france4\/0\/img.jpg",
		"genre_simplifie": "S\u00e9rie d'animation",
		"image_small": "\/image\/referentiel_emissions\/184533015\/1533577896\/80\/0\/france4\/0\/img.jpg",
		"ts_diffusion_utc": "1533475200",
		"bureau_regional": "",
		"acteurs": "",
		"date_diffusion": "2018-08-05T15:20",
		"invites": "",
		"url_image_racine": "\/image\/referentiel_emissions\/184533015\/1533577896",
		"code_programme": "inspecteur_gadget_france_4",
		"lsf": "0",
		"soustitre": "A fond de train",
		"titre": "Inspecteur Gadget",
		"id_programme": "24593",
		"genre": "S\u00e9rie d'animation",
		"id_collection": "15202833",
		"image_100": "\/image\/referentiel_emissions\/184533015\/1533577896\/100\/0\/france4\/0\/img.jpg",
		"audiodescription": "0",
		"image_300": "\/image\/referentiel_emissions\/184533015\/1533577896\/300\/0\/france4\/0\/img.jpg",
		"hashtag": "",
		"accroche": "Alors qu'il regarde ses trains, le Docteur Gang r\u00eave de mettre la main sur la derni\u00e8re invention de Sire Owen Barnstormer : le Planeur Express, un train \u00e0 la fine pointe de la technologie...",
		"genre_filtre": "seriedanimation",
		"titre_programme": "Inspecteur Gadget - France 4",
		"presentateurs": "",
		"accroche_programme": "Retrouvez les nouvelles aventures de l'inspecteur gadget",
		"csa_nom_long": "Tous publics",
		"image_large": "\/image\/referentiel_emissions\/184533015\/1533577896\/480\/0\/france4\/0\/img.jpg",
		"id_diffusion": "184533015",
		"rubrique": "jeunesse",
		"chaine_id": "france4",
		"nationalite": "canado-am\u00e9ricaine",
		"id_emission": "96045852",
		"chaine_label": "France 4",
		"csa_code": "TP",
		"saison": "1",
		"episode": "",
		"multilingue": "0",
		"format": "S\u00e9rie",
		"duree": "10",
		"extension_image": "jpg"
	}
	`
	v := &pluzzEmission{}
	d := json.NewDecoder(strings.NewReader(js))
	err := d.Decode(v)
	if err != nil {
		t.Errorf("Can't decode structure: %v", err)
		return
	}

}

func TestStructurePluzzList(t *testing.T) {
	const js = `{
		"query": {
			"string": "\/pluzz\/liste\/type\/replay\/nb\/999\/debut\/0",
			"params": {
				"support": "web",
				"version": "v1.3",
				"mode": "liste",
				"type": "replay",
				"nb": "999",
				"debut": "0",
				"tri": "datediff",
				"sens": "desc"
			},
			"tps_exec": 3.2271540164948,
			"memory_peak": 12730888,
			"generate-date": "Tue, 07 Aug 2018 17:29:38 +0200"
		},
		"reponse": {
			"debut": 0,
			"nb": 999,
			"total": 2855,
			"total_diffusions": 2855,
			"emissions": [
				{
					"nb_vues": "0",
					"volonte_replay": "1",
					"replay": "1",
					"mandat_duree": "7",
					"etranger": "0",
					"url": "\/videos\/ninjago_f4_saison8_ep10_,184747251.html",
					"soustitrage": "1",
					"recurrent": "1",
					"url_video_sitemap": "",
					"temps_restant": "603953",
					"duree_reelle": "1267",
					"nb_diffusion": "5",
					"oas_sitepage": "france4\/jeunesse\/ninjagof4",
					"realisateurs": "Trylle Vilstrup",
					"image_200": "\/image\/referentiel_emissions\/184747251\/1533643538\/200\/0\/france4\/0\/img.jpg",
					"image_medium": "\/image\/referentiel_emissions\/184747251\/1533643538\/214\/0\/france4\/0\/img.jpg",
					"genre_simplifie": "S\u00e9rie d'animation",
					"image_small": "\/image\/referentiel_emissions\/184747251\/1533643538\/80\/0\/france4\/0\/img.jpg",
					"ts_diffusion_utc": "1533653700",
					"bureau_regional": "",
					"acteurs": "",
					"date_diffusion": "2018-08-07T16:55",
					"invites": "",
					"url_image_racine": "\/image\/referentiel_emissions\/184747251\/1533643538",
					"code_programme": "ninjago_f4",
					"lsf": "0",
					"soustitre": "Petite Ninjago, gros ennuis",
					"titre": "Ninjago",
					"id_programme": "16665",
					"genre": "S\u00e9rie d'animation",
					"id_collection": "8280977",
					"image_100": "\/image\/referentiel_emissions\/184747251\/1533643538\/100\/0\/france4\/0\/img.jpg",
					"audiodescription": "0",
					"image_300": "\/image\/referentiel_emissions\/184747251\/1533643538\/300\/0\/france4\/0\/img.jpg",
					"hashtag": "",
					"accroche": "Garmadon entreprend de d\u00e9truire Ninjago \u00e0 l'aide d'un gigantesque robot. Les ninjas soignent Lloyd, gravement bless\u00e9, et d\u00e9couvrent qu'il a perdu ses pouvoirs...",
					"genre_filtre": "seriedanimation",
					"titre_programme": "Ninjago - F4",
					"presentateurs": "",
					"accroche_programme": "Les Ninjago attaquent France 4\r\nPour r\u00e9cup\u00e9rer les 4 armes d'or sacr\u00e9es,les Ninjas Kai,Cole,Jay et Zane devront parcourir les terres de Ninjago, ma\u00eetriser le puissant pouvoir du Spinjitzu et affronter le mal\u00e9fique Lord Garmadon et son arm\u00e9e de squelettes.",
					"csa_nom_long": "Tous publics",
					"image_large": "\/image\/referentiel_emissions\/184747251\/1533643538\/480\/0\/france4\/0\/img.jpg",
					"id_diffusion": "184747251",
					"rubrique": "jeunesse",
					"chaine_id": "france4",
					"nationalite": "am\u00e9ricaine",
					"id_emission": "128869967",
					"chaine_label": "France 4",
					"csa_code": "TP",
					"saison": "8",
					"episode": "10",
					"multilingue": "0",
					"format": "S\u00e9rie",
					"duree": "20",
					"extension_image": "jpg"
				},
				{
					"nb_vues": "0",
					"volonte_replay": "1",
					"replay": "1",
					"mandat_duree": "8",
					"etranger": "0",
					"url": "\/videos\/personne_n_y_avait_pense_,182273666.html",
					"soustitrage": "0",
					"recurrent": "",
					"url_video_sitemap": "",
					"temps_restant": "690922",
					"duree_reelle": "1973",
					"oas_sitepage": "france3\/jeu\/personnenyavaitpense",
					"realisateurs": "",
					"image_200": "\/image\/referentiel_emissions\/182273666\/1533643514\/200\/0\/france3\/0\/img.jpg",
					"image_medium": "\/image\/referentiel_emissions\/182273666\/1533643514\/214\/0\/france3\/0\/img.jpg",
					"genre_simplifie": "Jeu",
					"ts_diffusion_utc": "1533653100",
					"image_small": "\/image\/referentiel_emissions\/182273666\/1533643514\/80\/0\/france3\/0\/img.jpg",
					"acteurs": "",
					"bureau_regional": "",
					"invites": "",
					"date_diffusion": "2018-08-07T16:45",
					"url_image_racine": "\/image\/referentiel_emissions\/182273666\/1533643514",
					"code_programme": "personne_n_y_avait_pense",
					"soustitre": "",
					"lsf": "0",
					"titre": "Personne n'y avait pens\u00e9 !",
					"genre": "Jeu",
					"id_collection": "6622145",
					"id_programme": "7113",
					"image_100": "\/image\/referentiel_emissions\/182273666\/1533643514\/100\/0\/france3\/0\/img.jpg",
					"audiodescription": "0",
					"image_300": "\/image\/referentiel_emissions\/182273666\/1533643514\/300\/0\/france3\/0\/img.jpg",
					"hashtag": "#PNYAP",
					"genre_filtre": "jeu",
					"accroche": "Sous la houlette de Cyril F\u00e9raud, trois \u00e9quipes de candidats, r\u00e9parties en bin\u00f4mes, s'affrontent au cours de trois manches pour parvenir en finale. A chaque question, les candidats doivent trouver la r\u00e9ponse la moins \u00e9vidente bas\u00e9e sur les r\u00e9ponses d'un panel de 100 personnes pr\u00e9c\u00e9demment questionn\u00e9es. Pour remporter la cagnotte, le dernier bin\u00f4me encore en lice doit r\u00e9pondre \u00e0 une ultime question et proposer trois r\u00e9ponses dont l'une doit \u00eatre celle \u00e0 laquelle personne n'avait pens\u00e9. Qu'ils remportent ou non la finale, les candidats remettent leur titre en jeu dans l'\u00e9mission suivante.",
					"titre_programme": "Personne n'y avait pens\u00e9 !",
					"presentateurs": "Cyril F\u00e9raud",
					"csa_nom_long": "Tous publics",
					"accroche_programme": "Sous la houlette de Cyril F\u00e9raud, quatre bin\u00f4mes tentent de fournir \u00e0 des questions la r\u00e9ponse la moins \u00e9vidente possible. Celle-ci se fonde sur les r\u00e9sultats d'un panel de 100 personnes pr\u00e9alablement interrog\u00e9es. Si l'\u00e9quipe en finale ne remporte pas l'ultime \u00e9preuve...",
					"image_large": "\/image\/referentiel_emissions\/182273666\/1533643514\/480\/0\/france3\/0\/img.jpg",
					"id_diffusion": "182273666",
					"rubrique": "jeu",
					"nationalite": "",
					"chaine_id": "france3",
					"saison": "",
					"csa_code": "TP",
					"chaine_label": "France 3",
					"id_emission": "133688629",
					"episode": "",
					"multilingue": "0",
					"format": "Autre",
					"duree": "40",
					"extension_image": "jpg"
				}
			]
		}
	}`
	v := &pluzzList{}
	d := json.NewDecoder(strings.NewReader(js))
	err := d.Decode(v)
	if err != nil {
		t.Errorf("Can't decode structure: %v", err)
		return
	}
}

func TestReadFile(t *testing.T) {
	f := filepath.Join("testdata", "list10.json")

	r, err := os.Open(f)
	if err != nil {
		t.Error(err)
	}
	defer r.Close()

	d := json.NewDecoder(r)
	l := &pluzzList{}
	err = d.Decode(l)
	if err != nil {
		t.Errorf("Can't decode structure: %v", err)
		return
	}

}
