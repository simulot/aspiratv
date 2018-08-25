package francetv

import (
	"github.com/simulot/aspiratv/parsers/jsonparser"
)

// JSON structure for availble show list
// Commented fields are not used, this will save some work for the JSON decoder
type pluzzList struct {
	Reponse struct {
		// Debut           int             `json:"debut"`
		Emissions []pluzzEmission `json:"emissions"`
		// Nb              int             `json:"nb"`
		// Total           int             `json:"total"`
		// TotalDiffusions int             `json:"total_diffusions"`
	} `json:"reponse"`
}

type pluzzEmission struct {
	Accroche string `json:"accroche"`
	// AccrocheProgramme string `json:"accroche_programme"`
	// Acteurs           string `json:"acteurs"`
	// Audiodescription  string `json:"audiodescription"`
	// BureauRegional    string `json:"bureau_regional"`
	ChaineID string `json:"chaine_id"`
	// ChaineLabel       string `json:"chaine_label"`
	CodeProgramme string `json:"code_programme"`
	// CsaCode           string `json:"csa_code"`
	// CsaNomLong        string `json:"csa_nom_long"`
	DateDiffusion string `json:"date_diffusion"`
	// Duree         string  `json:"duree"`
	DureeReelle jsonparser.Seconds `json:"duree_reelle"`
	Episode     string             `json:"episode"`
	// Etranger          string `json:"etranger"`
	// ExtensionImage    string `json:"extension_image"`
	// Format            string `json:"format"`
	// Genre             string `json:"genre"`
	GenreFiltre string `json:"genre_filtre"`
	// GenreSimplifie    string `json:"genre_simplifie"`
	// Hashtag           string `json:"hashtag"`
	IDCollection string `json:"id_collection"`
	IDDiffusion  string `json:"id_diffusion"`
	IDEmission   string `json:"id_emission"`
	IDProgramme  string `json:"id_programme"`
	// Image100          string `json:"image_100"`
	// IMyJsonNamemage200          string `json:"image_200"`
	// IMyJsonNamemage300          string `json:"image_300"`
	ImageLarge string `json:"image_large"`
	// IMyJsonNamemageMedium       string `json:"image_medium"`
	// IMyJsonNamemageSmall        string `json:"image_small"`
	// IMyJsonNamenvites           string `json:"invites"`
	// LMyJsonNamesf               string `json:"lsf"`
	// MMyJsonNameandatDuree       string `json:"mandat_duree"`
	// MMyJsonNameultilingue       string `json:"multilingue"`
	// Nationalite       string `json:"nationalite"`
	// NbDiffusion       string `json:"nb_diffusion"`
	// NbVues            string `json:"nb_vues"`
	OasSitepage string `json:"oas_sitepage"`
	// Presentateurs     string `json:"presentateurs"`
	// Realisateurs      string `json:"realisateurs"`
	// Recurrent         string `json:"recurrent"`
	// Replay            string `json:"replay"`
	Rubrique string `json:"rubrique"`
	Saison   string `json:"saison"`
	// Soustitrage       string `json:"soustitrage"`
	Soustitre string `json:"soustitre"`
	// TempsRestant      string `json:"temps_restant"`
	Titre          string            `json:"titre"`
	TitreProgramme string            `json:"titre_programme"`
	TsDiffusionUtc jsonparser.TSUnix `json:"ts_diffusion_utc"`
	URL            string            `json:"url"`
	// URLImageRacine    string `json:"url_image_racine"`
	// URLVideoSitemap   string `json:"url_video_sitemap"`
	// VolonteReplay     string `json:"volonte_replay"`
}

// Show detail
type infoOeuvre struct {
	// Audiodescription bool        `json:"audiodescription"`
	// Chaine           string      `json:"chaine"`
	// CodeProgramme    string      `json:"code_programme"`
	// Credit           interface{} `json:"credit"`
	// CuePoints        interface{} `json:"cue_points"`
	// Diffusion        struct {
	// DateDebut string `json:"date_debut"`
	// Timestamp int    `json:"timestamp"`
	// } `json:"diffusion"`
	// Direct interface{} `json:"direct"`
	// Droit  struct {
	// Csa  string `json:"csa"`
	// Type string `json:"type"`
	// } `json:"droit"`
	// Duree                string        `json:"duree"`
	Episode int `json:"episode"`
	// Genre                string        `json:"genre"`
	// GenrePluzz           string        `json:"genre_pluzz"`
	// GenrePluzzAntenne    string        `json:"genre_pluzz_antenne"`
	// ID                   string        `json:"id"`
	// IDAedra              string        `json:"id_aedra"`
	// IDEmissionPlurimedia int           `json:"id_emission_plurimedia"`
	// IDTaxo               interface{}   `json:"id_taxo"`
	// Image                string        `json:"image"`
	ImageSecure string `json:"image_secure"` // Thumbnail
	// Indexes              []interface{} `json:"indexes"`
	// Lectures             struct {
	// 	ID         interface{} `json:"id"`
	// 	NbLectures int         `json:"nb_lectures"`
	// } `json:"lectures"`
	// LecturesGroupes  []interface{} `json:"lectures_groupes"`
	// MediamatIDChaine int           `json:"mediamat_id_chaine"`
	// Ordre            interface{}   `json:"ordre"`
	// Personnes        []struct {
	// 	Fonctions []string `json:"fonctions"`
	// 	Nom       string   `json:"nom"`
	// 	Prenom    string   `json:"prenom"`
	// } `json:"personnes"`
	// RealDuration      int           `json:"real_duration"`
	// RefSource         string        `json:"ref_source"`
	// Region            interface{}   `json:"region"`
	Saison int `json:"saison"`
	// SemaineDiffusion  interface{}   `json:"semaine_diffusion"`
	// Sequences         []interface{} `json:"sequences"`
	SousTitre string `json:"sous_titre"`
	// Spritesheet       []string      `json:"spritesheet"`
	// SpritesheetSecure []string      `json:"spritesheet_secure"`
	// Subtitles         []struct {
	// 	Format string `json:"format"`
	// 	Type   string `json:"type"`
	// 	URL    string `json:"url"`
	// } `json:"subtitles"`
	Synopsis string `json:"synopsis"`
	// TagOAS          interface{} `json:"tag_OAS"`
	// TexteDiffusions string      `json:"texte_diffusions"`
	Titre string `json:"titre"`
	// Type  string `json:"type"`
	// URLGuidetv      interface{} `json:"url_guidetv"`
	// URLReference string `json:"url_reference"`
	// URLSite string `json:"url_site"`
	Videos []struct {
		// Drm bool `json:"drm"`
		// DroitsLectureHorsConnexion bool     `json:"droits_lecture_hors_connexion"`
		// Embed                      bool     `json:"embed"`
		Format string `json:"format"`
		// Geoblocage                 []string `json:"geoblocage"`
		// PlagesOuverture            []struct {
		// 	Debut     int         `json:"debut"`
		// 	Direct    bool        `json:"direct"`
		// 	Fin       int         `json:"fin"`
		// 	Startover interface{} `json:"startover"`
		// } `json:"plages_ouverture"`
		// Statut    string `json:"statut"`
		URL string `json:"url"`
		// URLSecure string `json:"url_secure"`
	} `json:"videos"`
	// Votes interface{} `json:"votes"`
}
