package frontend

import "github.com/maxence-charriere/go-app/v9/pkg/app"

type CreditsPage struct {
	app.Compo
}

func newCreditPage(initialValue interface{}) app.Composer {
	return &CreditsPage{}
}

func (c *CreditsPage) OnMount(ctx app.Context) {
	ctx.Handle("BeforeMoving", c.beforeMoving)
}

func (c *CreditsPage) beforeMoving(ctx app.Context, action app.Action) {
	ctx.NewAction("PushHistory").Tag("page", PageCredits.String()).Tag("title", "Crédits").Post()

}
func (c *CreditsPage) Render() app.UI {
	MyAppState.CurrentPage = PageCredits

	return app.Section().
		Class("section").
		Body(
			app.H1().Class("title is-1").Text("Auteur"),
			app.Text("L'application AspiraTV est développée par Jean-François CASSAN."),
		)
}
