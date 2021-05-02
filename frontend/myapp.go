package frontend

import (
	"log"
	"time"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/simulot/aspiratv/frontend/bulma"
	"github.com/simulot/aspiratv/models"
)

// MyApp component draw de application banner and menus
type MyApp struct {
	app.Compo
	UpdateAvailable bool
}

func (c *MyApp) OnMount(ctx app.Context) {
}

func (c *MyApp) OnAppUpdate(ctx app.Context) {
	c.UpdateAvailable = ctx.AppUpdateAvailable // Reports that an app update is available.
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
		app.If(c.UpdateAvailable, app.Button().Text("Mettre à jour").OnClick(c.onUpdateClick)),
	)
}

func AppPageRender(pages ...app.UI) app.UI {
	return app.Div().
		Class("container").
		Body(
			&LoadSettings{},
			NewToastContainer(),
			app.Div().
				Class("columns").
				Body(
					&MyApp{},
					app.Div().
						Class("column").
						Body(
							app.Range(pages).
								Slice(func(i int) app.UI {
									return pages[i]
								}),
						),
				),
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

type ToastContainer struct {
	app.Compo
	unsubscribe func()
}

func NewToastContainer() *ToastContainer {
	return &ToastContainer{}
}

func (c *ToastContainer) OnMount(ctx app.Context) {
	c.unsubscribe = MyAppState.Drawer.OnChange(func() {
		ctx.Dispatch(func(ctx app.Context) {
			ns := MyAppState.Drawer.Notifications()
			for i := 0; i < len(ns); i++ {
				log.Printf("-- %d: %s", i, ns[i].Text)
			}
		})
	})
}

func (c *ToastContainer) OnDismount() {
	if c.unsubscribe != nil {
		c.unsubscribe()
	}
}

func (c *ToastContainer) Render() app.UI {
	ns := MyAppState.Drawer.Notifications()
	return app.Div().
		Class("toast-container column is-4 is-offset-8").
		Body(
			app.Range(ns).
				Slice(func(i int) app.UI {
					return NewToast(ns[i])
				}),
		)
}

type Toast struct {
	app.Compo
	Dismiss func()
	models.Notification
}

func NewToast(n models.Notification) *Toast {
	c := &Toast{
		Dismiss:      func() { MyAppState.Drawer.Dismiss(n) },
		Notification: n,
	}
	if n.Type != models.NotificationError {
		time.AfterFunc(4*time.Second, c.Dismiss)
	}
	return c
}

func (c *Toast) Render() app.UI {
	class := map[models.NotificationType]string{
		models.NotificationError:   "is-danger",
		models.NotificationInfo:    "is-info",
		models.NotificationSuccess: "is-success",
		models.NotificationWarning: "is-warning",
	}[c.Notification.Type]

	n := bulma.NewNotification().Class(class).Text(c.Notification.Text)
	if c.Type == models.NotificationError {
		n.OnDelete(c.Dismiss)
	}
	return n
}
