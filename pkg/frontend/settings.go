package frontend

import (
	"fmt"
	"log"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/simulot/aspiratv/pkg/models"
)

const (
	labelSettings               = "Réglages"
	labelLibraryPath            = "Chemin de la bibliothèque"
	labelLibraryPathPlaceHolder = "donnez un chemin"
	labelSubmit                 = "Envoyer"
	labelCancel                 = "Annuler"
)

type Settings struct {
	app.Compo
	Settings models.Settings

	i int
}

func (c *Settings) OnMount(ctx app.Context) {
	if !MyAppState.StateReady {
		ctx.Async(func() {

			log.Printf("Settings is waiting")
			<-MyAppState.Ready
			c.Update()
		})
	}
}

func (c *Settings) Render() app.UI {
	MyAppState.CurrentPage = PageSettings
	return AppPageRender(
		app.H1().
			Class("title is-1").
			Text(labelSettings),
		app.Div().
			Class("field").
			Body(
				app.Label().
					Class("label").
					Text(labelLibraryPath),
				app.Div().
					Class("control").
					Body(
						app.Input().
							Class("input").
							Type("text").
							Placeholder(labelLibraryPathPlaceHolder).
							AutoFocus(true).
							Value(MyAppState.Settings.LibraryPath).
							OnChange(c.ValueTo(&c.Settings.LibraryPath))),
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

func (c *Settings) messageError(ctx app.Context, e app.Event) {
	c.i++
	MyAppState.Dispatch.Publish(models.NewNotification(fmt.Sprintf("Message %d !", c.i), models.NotificationError))
}
func (c *Settings) messageSuccess(ctx app.Context, e app.Event) {
	c.i++
	MyAppState.Dispatch.Publish(models.NewNotification(fmt.Sprintf("Message %d !", c.i), models.NotificationSuccess))
}

func (c *Settings) submit(ctx app.Context, e app.Event) {
	s, err := MyAppState.API.SetSettings(ctx, MyAppState.Settings)
	if err != nil {
		MyAppState.Dispatch.Publish(models.NewNotification(err.Error(), models.NotificationError))
		return
	}
	MyAppState.Settings = s
	MyAppState.Dispatch.Publish(models.NewNotification("Réglages enregistrés", models.NotificationSuccess))
}

func (c *Settings) cancel(ctx app.Context, e app.Event) {
	s, err := MyAppState.GetSettings(ctx)
	if err != nil {
		MyAppState.Dispatch.Publish(models.NewNotification(err.Error(), models.NotificationError))
	}
	MyAppState.Settings = s
	c.Update()
}
