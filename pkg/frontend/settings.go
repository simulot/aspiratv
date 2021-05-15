package frontend

import (
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"time"

	"github.com/kr/pretty"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/simulot/aspiratv/pkg/frontend/bulma"
	"github.com/simulot/aspiratv/pkg/models"
)

const (
	labelSettings               = "Réglages"
	labelLibraryPath            = "Dossier contenant la bibliothèque sur le serveur"
	labelShowPathTemplate       = "Modèle pour le nom de dossier pour la série, l'émission"
	labelSeasonPathTemplate     = "Modèle pour le nom de dossier pour la saison"
	labelShowNameTemplate       = "Modèle pour le nom du fichier vidéo"
	labelTypePathNaming         = "Modèle de nom de fichier"
	labelLibraryPathPlaceHolder = "donnez un chemin"
	labelSubmit                 = "Envoyer"
	labelCancel                 = "Annuler"
	labelSubFolder              = "Sous-dossier"
	labelSubFolderPlaceHolder   = "dossier dans la librairie"
)

var labelsTypePathNaming = map[models.PathNamingType]string{
	// models.PathTypeMovie: "Style Film",
	// models.PathTypeCustom:     "Style Défini par l'utilisateur",
	models.PathTypeCollection: "Style Collection",
	models.PathTypeSeries:     "Style Séries",
	models.PathTypeTVShow:     "Style Magazine",
}
var SampleMedia = map[models.PathNamingType]models.MediaInfo{
	models.PathTypeCollection: {
		Show:     "Science et découverte",
		Type:     models.TypeCollection,
		Aired:    time.Date(2020, 12, 11, 0, 0, 0, 0, time.Local).Truncate(0),
		Title:    "Les pouvoirs du cerveau - Notre intelligence dévoilé",
		Provider: "artetv",
		Channel:  "Arte",
	},
	models.PathTypeSeries: {
		Show:     "50 nuances de Grecs",
		Type:     models.TypeSeries,
		Aired:    time.Date(2020, 8, 28, 0, 0, 0, 0, time.Local).Truncate(0),
		Season:   2,
		Episode:  3,
		Title:    "Scène de ménage",
		Provider: "artetv",
		Channel:  "Arte",
	},
	models.PathTypeTVShow: {
		Show:     "C dans l'air",
		Type:     models.TypeTVShow,
		Aired:    time.Date(2021, 5, 12, 0, 0, 0, 0, time.Local).Truncate(0),
		Year:     2021,
		Title:    "Vaccins : les labos vont-ils lâcher leurs brevets ?",
		Provider: "francetv",
		Channel:  "France 5",
	},
}

type Settings struct {
	app.Compo
	Settings    models.Settings
	LibraryPath *string
	v, h        string

	i int
}

func (c *Settings) OnMount(ctx app.Context) {
	<-MyAppState.Ready
	c.Settings = MyAppState.Settings
}

// func (c *Settings) OnPreRender(ctx app.Context) {
// 	<-MyAppState.Ready
// 	c.Settings = MyAppState.Settings
// 	log.Printf("%#v", c.Settings)
// }

func (c *Settings) Render() app.UI {
	MyAppState.CurrentPage = PageSettings
	c.Settings = MyAppState.Settings
	c.LibraryPath = &c.Settings.LibraryPath
	return AppPageRender(
		app.H1().
			Class("title is-1").
			Text(labelSettings),
		bulma.NewTextField(&c.Settings.LibraryPath, labelLibraryPath, labelLibraryPathPlaceHolder).WithOnInput(func(ctx app.Context, v string) (string, error) {
			c.Settings.LibraryPath = v
			return v, nil
		}),
		NewPathSettings("Séries", c.Settings.SeriesSettings, c.Settings.LibraryPath),
		NewPathSettings("Émissions", c.Settings.TVShowsSettings, c.Settings.LibraryPath),
		NewPathSettings("Collections", c.Settings.CollectionSettings, c.Settings.LibraryPath),
		app.Div().
			Class("field is-grouped").
			Body(
				app.Div().
					Class("control").
					Body(
						app.Button().
							Class("button is-link").
							OnClick(c.submit).
							Text(labelSubmit),
					),
				app.Div().
					Class("control").
					Body(
						app.Button().
							Class("button is-link").
							OnClick(c.cancel).
							Text(labelCancel),
					),
			),
		app.Div().
			Class("field is-grouped").
			Body(
				app.Div().
					Class("control").
					Body(
						app.Button().
							Class("button is-link").
							OnClick(c.messageError).
							Text("Message Erreur!"),
					),
				app.Button().
					Class("button is-link").
					OnClick(c.messageSuccess).
					Text("Message Succés!"),
			),
	)
}

type PathSetting struct {
	app.Compo
	Legend      string
	PathSetting *models.PathSettings
	LibraryPath string
	Folder      *string
	sample      string
}

func NewPathSettings(legend string, s *models.PathSettings, libraryPath string) *PathSetting {
	return &PathSetting{
		PathSetting: s,
		Legend:      legend,
		LibraryPath: libraryPath,
		Folder:      &s.Folder,
	}
}

func (c *PathSetting) Render() app.UI {
	s := bulma.NewSelectField(labelTypePathNaming).WhitOnInput(func(ctx app.Context, selected string) (string, error) {
		option, _ := strconv.Atoi(selected)
		c.PathSetting.PathNaming = models.PathNamingType(option)
		return selected, nil
	})
	for k := range SampleMedia {
		s.WithOption(
			strconv.Itoa(int(k)),
			labelsTypePathNaming[k],
			k == c.PathSetting.PathNaming,
		)
	}

	namer := models.DefaultFileNamer[c.PathSetting.PathNaming]
	m := SampleMedia[c.PathSetting.PathNaming]
	c.sample = filepath.Join(c.LibraryPath, *c.Folder, namer.ShowPathString(m), namer.SeasonPathString(m), namer.MediaFileNameString(m))

	return app.Section().Body(
		app.H2().Class("title is-2").Text(c.Legend),
		bulma.NewTextField(&c.PathSetting.Folder, labelSubFolder, labelSubFolderPlaceHolder),
		app.Div().Class("field").Body(
			app.Label().Class("label").Text(labelSubFolder),
			s,
		),
		app.P().Body(
			app.Text("Exemple: "),
			app.Text(c.sample),
		),
	)
}

func (c *Settings) messageError(ctx app.Context, e app.Event) {
	c.i++
	MyAppState.Dispatch.Publish(models.NewMessage(fmt.Sprintf("Message %d !", c.i)).SetStatus(models.StatusError).SetPinned(true))
}
func (c *Settings) messageSuccess(ctx app.Context, e app.Event) {
	c.i++
	MyAppState.Dispatch.Publish(models.NewMessage(fmt.Sprintf("Message %d !", c.i)).SetStatus(models.StatusSuccess))
}

func (c *Settings) submit(ctx app.Context, e app.Event) {
	log.Printf("Settings: %# v", pretty.Formatter(c.Settings))
	s, err := MyAppState.API.SetSettings(ctx, c.Settings)
	if err != nil {
		MyAppState.Dispatch.Publish(models.NewMessage(err.Error()).SetStatus(models.StatusError).SetPinned(true))
		return
	}
	MyAppState.Settings = s
	MyAppState.Dispatch.Publish(models.NewMessage("Réglages enregistrés").SetStatus(models.StatusSuccess))
}

func (c *Settings) cancel(ctx app.Context, e app.Event) {
	s, err := MyAppState.GetSettings(ctx)
	if err != nil {
		MyAppState.Dispatch.Publish(models.NewMessage(err.Error()).SetStatus(models.StatusError).SetPinned(true))
	}
	MyAppState.Settings = s
	c.Update()
}
