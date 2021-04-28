package frontend

import (
	"log"
	"sync"
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

func AppPageRender(pages ...app.UI) app.UI {
	return app.Div().
		Class("container").
		Body(
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
	messages []*appMessage
	sync.RWMutex
	unsubscribe func()
}

func NewToastContainer() *ToastContainer {
	return &ToastContainer{}
}

func (c *ToastContainer) OnMount(ctx app.Context) {
	log.Printf("ToastContainer.OnMount")
	c.unsubscribe = MyAppState.Messages.Subscribe(func(m appMessage) {
		ctx.Dispatch(func(ctx app.Context) {
			c.AddMessage(ctx, m)
		})
	})
}

func (c *ToastContainer) OnDismount() {
	c.unsubscribe()
}

func (c *ToastContainer) AddMessage(ctx app.Context, m appMessage) {
	c.Lock()
	defer c.Unlock()
	c.messages = append(c.messages, &m)
}

func (c *ToastContainer) Render() app.UI {
	c.RLock()
	defer c.RUnlock()
	return app.Div().
		Class("toast-container").
		Body(
			app.Range(c.messages).
				Slice(func(i int) app.UI {
					m := c.messages[i]
					return NewToast(func() { c.DismissMessage(m) }, m)
				}),
		)
}

func (c *ToastContainer) DismissMessage(m *appMessage) {
	c.Lock()
	defer c.Unlock()
	for i := 0; i < len(c.messages); i++ {
		if c.messages[i] == m {
			log.Printf("Dismiss %d", i)
			copy(c.messages[i:], c.messages[i+1:])      // Shift a[i+1:] left one index.
			c.messages[len(c.messages)-1] = nil         // Erase last element (write zero value).
			c.messages = c.messages[:len(c.messages)-1] // Truncate slice.
			c.Update()
			return
		}
	}
}

type Toast struct {
	app.Compo
	dismiss func()
	*appMessage
}

func NewToast(dismiss func(), m *appMessage) *Toast {
	c := &Toast{
		dismiss:    dismiss,
		appMessage: m,
	}
	if !m.Stay {
		time.AfterFunc(4*time.Second, dismiss)
	}
	return c
}

func (c *Toast) Render() app.UI {
	return app.Div().
		Class("toast").
		Body(
			app.Div().
				Class("notification").
				Class(StringIf(c.Class == "", "is-info", c.Class)).
				Body(
					app.If(c.Stay,
						app.Button().
							Class("delete").
							OnClick(c.Dismiss()),
					),
					c.Content,
				),
		)
}

func (c *Toast) Dismiss() func(ctx app.Context, e app.Event) {
	return func(ctx app.Context, e app.Event) {
		c.dismiss()
	}
}
