package frontend

/*
	Settings screen

	TODO:
		- Update path exemples when change library pass
		-

*/

import (
	"path/filepath"
	"strconv"
	"time"

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
	initialized bool
}

func (c *Settings) OnMount(ctx app.Context) {
	<-MyAppState.Ready
	c.Settings = MyAppState.Settings
	c.initialized = true
}

func (c *Settings) Render() app.UI {
	MyAppState.CurrentPage = PageSettings
	if !c.initialized {
		<-MyAppState.Ready
		c.Settings = MyAppState.Settings
		c.initialized = true
	}

	cSeries := NewPathSettings("Séries", c.Settings.SeriesSettings, &c.Settings.LibraryPath).WithOnInput(func(s models.PathSettings) {
		c.Settings.SeriesSettings = s
	})
	cTVShows := NewPathSettings("Émissions", c.Settings.TVShowsSettings, &c.Settings.LibraryPath).WithOnInput(func(s models.PathSettings) {
		c.Settings.TVShowsSettings = s
	})
	cCollections := NewPathSettings("Collections", c.Settings.CollectionSettings, &c.Settings.LibraryPath).WithOnInput(func(s models.PathSettings) {
		c.Settings.CollectionSettings = s
	})

	return AppPageRender(
		app.H1().
			Class("title is-1").
			Text(labelSettings),
		bulma.NewTextField(c.Settings.LibraryPath, labelLibraryPath, labelLibraryPathPlaceHolder).WithOnInput(func(v string) {
			c.Settings.LibraryPath = v
			cSeries.SetHelp()
			cTVShows.SetHelp()
			cCollections.SetHelp()
			c.Update()
		}),
		cSeries,
		cTVShows,
		cCollections,
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
	)
}

type PathSetting struct {
	app.Compo
	Legend      string
	pathSetting models.PathSettings
	libraryPath *string
	help        string
	selected    string
	onInputFn   func(models.PathSettings)
}

func NewPathSettings(legend string, s models.PathSettings, libraryPath *string) *PathSetting {
	return &PathSetting{
		pathSetting: s,
		Legend:      legend,
		libraryPath: libraryPath,
	}
}

func (c *PathSetting) WithOnInput(f func(models.PathSettings)) *PathSetting {
	c.onInputFn = f
	return c
}

func (c *PathSetting) onInput() {
	if c.onInputFn != nil {
		c.onInputFn(c.pathSetting)
	}
}

func (c *PathSetting) SetHelp() {
	namer := models.DefaultFileNamer[c.pathSetting.PathNaming]
	m := SampleMedia[c.pathSetting.PathNaming]

	c.help = filepath.Join(*c.libraryPath, c.pathSetting.Folder, namer.ShowPathString(m), namer.SeasonPathString(m), namer.MediaFileNameString(m))
}

func (c *PathSetting) Render() app.UI {
	s := bulma.NewSelectField(c.selected, labelTypePathNaming).WhitOnInput(func(selected string) {
		option, _ := strconv.Atoi(selected)
		c.pathSetting.PathNaming = models.PathNamingType(option)
		c.onInput()
		c.SetHelp()
	})
	for k := range SampleMedia {
		s.WithOption(
			strconv.Itoa(int(k)),
			labelsTypePathNaming[k],
			k == c.pathSetting.PathNaming,
		)
	}
	return app.Section().Body(
		app.H2().Class("title is-2").Text(c.Legend),
		bulma.NewTextField(c.pathSetting.Folder, labelSubFolder, labelSubFolderPlaceHolder).WithOnInput(func(v string) {
			c.pathSetting.Folder = v
			c.onInput()
			c.SetHelp()
		}),
		s,
		app.Div().Class("field").Body(
			app.P().Class("help").Body(
				app.Text("Exemple : "),
				app.Text(c.help),
			),
		),
	)
}

func (c *Settings) submit(ctx app.Context, e app.Event) {
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
