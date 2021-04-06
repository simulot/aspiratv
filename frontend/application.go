package frontend

import (
	"github.com/maxence-charriere/go-app/v8/pkg/app"
	"github.com/simulot/aspiratv/store"
)

type AspiraTVAppConfig struct {
	store store.Store
}

// TODO Get ride of global variable

var AspiraTVApp *AspiraTVAppConfig

func InitializeAspritaTVApp(c *AspiraTVAppConfig) {
	AspiraTVApp = c
}

// A component that describes a UI.
type MyApp struct {
	app.Compo

	// Field that reports whether an app update is available. False by default.
	updateAvailable bool
}

// OnAppUpdate satisfies the app.Updater interface. It is called when the app is
// updated in background.
func (a *MyApp) OnAppUpdate(ctx app.Context) {
	a.updateAvailable = ctx.AppUpdateAvailable // Reports that an app update is available.
	a.Update()                                 // Triggers UI update.
}

func (a *MyApp) Render() app.UI {
	return app.Main().Body(
		// Displays an Update button when an update is available.
		app.If(a.updateAvailable,
			app.Button().
				Text("Update!").
				OnClick(a.onUpdateClick),
		),
		app.H1().Text("A little app"),
		app.P().Text("That only display a text. updated sometime ago"),
		&hello{
			name: "World",
		},
	)
}

func (a *MyApp) onUpdateClick(ctx app.Context, e app.Event) {
	// Reloads the page to display the modifications.
	ctx.Reload()
}

type hello struct {
	app.Compo

	name string // Field where the username is stored
}

func (h *hello) Render() app.UI {
	return app.Div().Body(
		app.H1().Body(
			app.Text("Hello "),
			app.Text(h.name), // The name field used in the title
		),

		// The input HTML element that get the username.
		app.Input().
			Value(h.name).             // The name field used as current input value
			OnChange(h.OnInputChange), // The event handler that will store the username
	)
}

func (h *hello) OnInputChange(ctx app.Context, e app.Event) {
	h.name = ctx.JSSrc.Get("value").String() // Name field is modified
	h.Update()                               // Update the component UI
}
