package frontend

import (
	"github.com/maxence-charriere/go-app/v8/pkg/app"
)

// MyApp component draw de application banner and menus
type MyApp struct {
	app.Compo
	updateAvailable bool
}

func (c *MyApp) OnAppUpdate(ctx app.Context) {
	c.updateAvailable = ctx.AppUpdateAvailable // Reports that an app update is available.
	c.Update()                                 // Triggers UI update.
}

func (c *MyApp) onUpdateClick(ctx app.Context, e app.Event) {
	// Reloads the page to display the modifications.
	ctx.Reload()
}

func (c *MyApp) Render() app.UI {
	return app.Div().Class("column is-narrow").Body(
		&Logo{},
		&Menu{},
		app.If(c.updateAvailable, app.Button().Text("Mettre à jour").OnClick(c.onUpdateClick)),
	)
}

type Logo struct {
	app.Compo
}

func (c *Logo) Render() app.UI {
	return app.Div().Class("banner").Body(
		app.H1().Class("title").Text("AspiraTV"),
	)
}

type LandingPage struct {
	app.Compo
}

func (c *LandingPage) OnNav(ctx app.Context) {
	ctx.Navigate("/search")
}

func (c *LandingPage) Render() app.UI {
	return app.A().Href("/search").Text("Aller sur la page de recherche")
}
