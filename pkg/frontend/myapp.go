package frontend

import (
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type Menuitem struct {
	page        PageID
	icon        string
	label       string
	path        string
	selected    bool
	constructor func() app.Composer
}

var appMenus = []Menuitem{
	{
		page:        PageSearchOnLine,
		label:       "Chercher en ligne",
		path:        "/search",
		constructor: newSearch,
	},
	// {
	// 	PageLibrary,
	// 	"",
	// 	"Bibliothèque",
	// 	"/library",
	// },
	{
		page:        PageSubscriptions,
		label:       "Abonnements",
		path:        "/subscriptions",
		constructor: newSubscriptionPage,
	},

	{
		page:        PageSettings,
		label:       "Réglages",
		path:        "/settings",
		constructor: newSettingsPage,
	},
	{
		page:        PageCredits,
		label:       "Crédits",
		path:        "/credits",
		constructor: newCreditPage,
	},
}

// MyApp component draw de application banner and menus
type MyApp struct {
	app.Compo
	UpdateAvailable bool
	Notifications   bool

	ready         bool
	page          app.UI
	currentPageID PageID
}

func (a *MyApp) OnMount(ctx app.Context) {
	if a.currentPageID == 0 {
		a.currentPageID = PageSearchOnLine
	}
	a.GotoPage(a.currentPageID)
	a.ready = true
}

// func (c *MyApp) OnPreRender(ctx app.Context) {
// 	log.Printf("Settings is waiting")
// 	<-MyAppState.Ready

// }

// func (c *MyApp) OnMount(ctx app.Context) {
// 	if !MyAppState.StateReady {
// 		ctx.Async(func() {
// 			log.Printf("Settings is waiting")
// 			<-MyAppState.Ready
// 			c.Update()
// 		})
// 	}
// }

func (a *MyApp) OnAppUpdate(ctx app.Context) {
	a.UpdateAvailable = ctx.AppUpdateAvailable() // Reports that an app update is available.
	a.Update()                                   // Triggers UI update.
}

func (a *MyApp) onUpdateClick(ctx app.Context, e app.Event) {
	// Reloads the page to display the modifications.
	ctx.Reload()
}

func (a *MyApp) Render() app.UI {
	return app.Div().
		Class("container").
		Body(
			&LoadSettings{},
			NewMessagesContainer(),
			app.Div().
				Class("columns").
				Body(
					app.Div().
						Class("column is-narrow").
						Body(
							&Logo{},
							a.RenderMenus(),
							app.If(a.UpdateAvailable, app.Button().
								Text("Mettre à jour").
								OnClick(a.onUpdateClick),
							),
						),
					app.Div().
						Class("column").
						Body(
							a.page,
						),
				))
}

func (a *MyApp) RenderMenus() app.UI {
	return app.Div().Class("menu").Body(
		app.Ul().Class("menu-list").Body(
			app.Range(appMenus).Slice(func(i int) app.UI {
				item := appMenus[i]
				return app.Li().Body(
					app.A().
						Class(StringIf(item.page == a.currentPageID, "is-active", "")).
						Text(item.label).
						OnClick(a.menuClick(item.page)),
				)
			}),
		),
	)
}

func (a *MyApp) menuClick(page PageID) app.EventHandler {
	return func(ctx app.Context, e app.Event) {
		a.GotoPage(page)
	}
}

func (a *MyApp) GotoPage(page PageID) {
	for i := range appMenus {
		if appMenus[i].page == page {
			appMenus[i].selected = true
			a.page = appMenus[i].constructor()
			a.currentPageID = page
		} else {
			appMenus[i].selected = false
		}
	}
}

type Logo struct {
	app.Compo
}

func (c *Logo) Render() app.UI {
	return app.Div().Class("banner").Body(
		app.H1().Class("title").Text("AspiraTV"),
	)
}
