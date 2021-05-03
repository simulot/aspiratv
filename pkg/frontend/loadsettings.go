package frontend

import (
	"log"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type LoadSettings struct {
	app.Compo
}

func (l *LoadSettings) OnMount(ctx app.Context) {
	if !MyAppState.StateReady {
		ctx.Async(func() {
			log.Printf("LoadSettings is waiting")
			<-MyAppState.Ready
			log.Printf("LoadSettings done waiting")
			l.Update()
		})
	}
}

func (l *LoadSettings) Render() app.UI {
	log.Printf("LoadSettings rendering %v", MyAppState.StateReady)
	return app.Div().ID("LOADING").
		Class("modal").
		Class(StringIf(!MyAppState.StateReady, "is-active", "")).
		Body(
			app.Div().Class("modal-background"),
			app.Div().Class("modal-content").
				Body(
					app.Div().Class("box").Body(
						app.P().Class("is-size-1 has-text-centered").Text("Initialisation ..."),
						app.P().Class("is-size-1 has-text-centered is-loading").Text("."),
					),
				),
		)
}
