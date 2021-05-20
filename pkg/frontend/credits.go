package frontend

import "github.com/maxence-charriere/go-app/v9/pkg/app"

type CreditsPage struct {
	app.Compo
}

func (c *CreditsPage) Render() app.UI {
	MyAppState.CurrentPage = PageCredits

	return app.Div().Class("container").Body(
		app.Div().Class("columns").Body(
			&MyApp{},
			app.Div().Class("column").Body(
				app.Section().Class("section").Body(
					app.H2().Class("title").Text("Auteur"),
					app.Text("L'application AspiraTV est développée par Jean-François CASSAN."),
				),
				app.Section().Class("section").Body(
					app.H2().Class("title").Text("Auteur"),
					app.Text("L'application AspiraTV est développée par Jean-François CASSAN."),
				),
			),
		),
	)
}
