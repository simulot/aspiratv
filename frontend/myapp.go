package frontend

import (
	"time"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
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

type appMessage struct {
	Class   string
	Stay    bool
	Content app.UI
}
type ToastContainer struct {
	app.Compo
	messages []*appMessage
}

func (c *ToastContainer) AddMessage(ctx app.Context, t string, class string, stay bool) {
	m := appMessage{
		Class:   class,
		Stay:    stay,
		Content: app.Text(t),
	}
	c.messages = append(c.messages, &m)
	if !stay {
		time.AfterFunc(4*time.Second, func() {
			ctx.Dispatch(func(ctx app.Context) {
				c.closeMessage(&m)
			})
		})
	}
}

func (c *ToastContainer) Render() app.UI {
	return app.Div().
		Class("toast-container").
		Body(
			app.Range(c.messages).
				Slice(func(i int) app.UI {
					return c.renderMessage(i)
				}),
		)
}

func (c *ToastContainer) renderMessage(i int) app.UI {
	m := c.messages[i]
	return app.Div().
		Class("toast").
		Body(
			app.Div().
				Class("notification").
				Class(StringIf(m.Class == "", "is-info", m.Class)).
				Body(
					app.If(m.Stay,
						app.Button().
							Class("delete").
							OnClick(c.toastDismiss(m)),
					),
					m.Content,
				),
		)
}

func (c *ToastContainer) toastDismiss(m *appMessage) func(ctx app.Context, e app.Event) {
	return func(ctx app.Context, e app.Event) {
		c.closeMessage(m)
		c.Update()
	}
}

func (c *ToastContainer) closeMessage(m *appMessage) {
	for i := 0; i < len(c.messages); i++ {
		if c.messages[i] == m {
			copy(c.messages[i:], c.messages[i+1:])      // Shift a[i+1:] left one index.
			c.messages[len(c.messages)-1] = nil         // Erase last element (write zero value).
			c.messages = c.messages[:len(c.messages)-1] // Truncate slice.
			return
		}
	}
}
