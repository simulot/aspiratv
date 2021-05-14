package frontend

import (
	"fmt"
	"log"

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
	labelLibraryPathPlaceHolder = "donnez un chemin"
	labelSubmit                 = "Envoyer"
	labelCancel                 = "Annuler"
)

type Settings struct {
	app.Compo
	Settings models.Settings
	v, h     string

	i int
}

func (c *Settings) Render() app.UI {
	MyAppState.CurrentPage = PageSettings
	c.Settings = MyAppState.Settings
	return AppPageRender(
		app.H1().
			Class("title is-1").
			Text(labelSettings),
		bulma.NewTextField(&c.Settings.LibraryPath, labelLibraryPath, labelLibraryPathPlaceHolder),
		NewPathSettings("Pour les séries", c.Settings.DefaultSeriesSettings),
		NewPathSettings("Pour les émissions", c.Settings.DefaultTVShowsSettings),
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
}

func NewPathSettings(legend string, s *models.PathSettings) *PathSetting {
	return &PathSetting{
		PathSetting: s,
		Legend:      legend,
	}
}

func (c *PathSetting) Render() app.UI {
	return app.Section().Body(
		app.H2().Class("title is-2").Text(c.Legend),
		bulma.NewTextField(&c.PathSetting.ShowPathTemplate, labelShowPathTemplate, labelLibraryPathPlaceHolder),
		bulma.NewTextField(&c.PathSetting.SeasonPathTemplate, labelSeasonPathTemplate, labelLibraryPathPlaceHolder),
		bulma.NewTextField(&c.PathSetting.MediaFileNameTemplate, labelShowNameTemplate, labelLibraryPathPlaceHolder),
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
