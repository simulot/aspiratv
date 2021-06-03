package frontend

import (
	"errors"
	"log"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type pageInfo struct {
	menuID      PageID
	constructor func(initialValue interface{}) app.Composer
}

var pageRegister = map[PageID]pageInfo{
	PageSearchOnLine: {
		PageSearchOnLine,
		newSearch,
	},
	PageSubscriptions: {
		PageSubscriptions,
		newSubscriptionListPage,
	},

	PageEditSubscrition: {
		PageSubscriptions,
		newSubscriptionPage,
	},

	PageSettings: {
		PageSettings,
		newSettingsPage,
	},
	PageCredits: {
		PageCredits,
		newCreditPage,
	},
}

type Menuitem struct {
	page     PageID
	icon     string
	label    string
	path     string
	selected bool
}

var appMenus = []Menuitem{
	{
		page:  PageSearchOnLine,
		label: "Chercher en ligne",
		path:  "/search",
	},
	// {
	// 	PageLibrary,
	// 	"",
	// 	"Bibliothèque",
	// 	"/library",
	// },
	{
		page:  PageSubscriptions,
		label: "Abonnements",
		path:  "/subscriptions",
	},

	{
		page:  PageSettings,
		label: "Réglages",
		path:  "/settings",
	},
	{
		page:  PageCredits,
		label: "Crédits",
		path:  "/credits",
	},
}

// MyApp component draw de application banner and menus
type MyApp struct {
	app.Compo
	UpdateAvailable bool
	Notifications   bool
	history         *applicationHistory

	ready         bool
	page          app.UI
	currentPageID PageID
}

func (a *MyApp) OnMount(ctx app.Context) {
	a.history = newApplicationHistory(ctx)

	ctx.Handle("GotoPage", a.GotoPageActionHandler)
	ctx.Handle("GotoBack", a.BackPageActionHandler)
	ctx.Handle("PushHistory", a.PushHistory)

	a.ready = true
	a.GotoPage(ctx, PageSearchOnLine, nil)
}

func (a *MyApp) GotoPageActionHandler(ctx app.Context, action app.Action) {
	pageID, _ := PageIDString(action.Tags.Get("page"))
	a.GotoPage(ctx, pageID, action.Value)

}

func (a *MyApp) BackPageActionHandler(ctx app.Context, action app.Action) {
	a.Back(ctx, action)
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
		GotoPage(ctx, page, nil)
	}
}

func (a *MyApp) GotoPage(ctx app.Context, pageID PageID, initialValue interface{}) {
	page := pageRegister[pageID]
	log.Printf("[NAVIGATION] Goto page %s", pageID)
	for i := range appMenus {
		if appMenus[i].page == page.menuID {
			appMenus[i].selected = true
			a.currentPageID = pageID
		} else {
			appMenus[i].selected = false
		}
	}
	a.page = page.constructor(initialValue)
}

func (a *MyApp) PushHistory(ctx app.Context, action app.Action) {
	log.Printf("[NAVIGATION] Push state for page %s", action.Tags.Get("page"))
	a.history.push(ctx, action)
}
func (a *MyApp) Back(ctx app.Context, action app.Action) {
	p, err := a.history.pop()
	if err != nil {
		log.Printf("[NAVIGATION] Can't go back: %s", err)
		return
	}

	log.Printf("[NAVIGATION] Back to page %s", p.page)
	a.GotoPage(ctx, p.page, p.state)
}

type Logo struct {
	app.Compo
}

func (c *Logo) Render() app.UI {
	return app.Div().Class("banner").Body(
		app.H1().Class("title").Text("AspiraTV"),
	)
}

type applicationHistory struct {
	history []PageState
}

func newApplicationHistory(ctx app.Context) *applicationHistory {
	h := applicationHistory{}
	return &h
}

func (h *applicationHistory) push(ctx app.Context, action app.Action) {
	s := actionToState(action)
	h.history = append(h.history, s)
}

func (h *applicationHistory) pop() (PageState, error) {
	if len(h.history) == 0 {
		return PageState{}, errors.New("No previous action")
	}
	s := h.history[len(h.history)-1]
	h.history = h.history[:len(h.history)-1]
	return s, nil
}
