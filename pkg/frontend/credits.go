package frontend

import "github.com/maxence-charriere/go-app/v9/pkg/app"

type Credits struct {
	app.Compo
}

func (c *Credits) Render() app.UI {
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
