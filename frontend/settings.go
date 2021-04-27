package frontend

import (
	"log"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/simulot/aspiratv/models"
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
}

func (c *Settings) OnMount(ctx app.Context) {
	ctx.Async(func() {
		c.getSettings(ctx)
		c.Update()
	})
}

func (c *Settings) getSettings(ctx app.Context) {
	s, err := MyAppState.s.GetSettings(ctx)
	if err != nil {
		log.Print("[SETTINGS] Cant get settings error: ", err)
		return
	}
	log.Printf("[SETTINGS] Get %#v", s)
	c.Settings = s
}

func (c *Settings) Render() app.UI {
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
							Value(c.Settings.LibraryPath).
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
	)
}

func (c *Settings) submit(ctx app.Context, e app.Event) {
	s, err := MyAppState.s.SetSettings(ctx, c.Settings)
	if err != nil {
		//TODO show error
		return
	}
	c.Settings = s

}

func (c *Settings) cancel(ctx app.Context, e app.Event) {
	c.getSettings(ctx)
	c.Update()
}
