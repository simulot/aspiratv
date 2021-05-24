package bulma

import "github.com/maxence-charriere/go-app/v9/pkg/app"

type ConfirmDialog struct {
	app.Compo
	title      string
	actionText string
	callBack   func(ctx app.Context, action bool) // Callback called when the action is confirmed
}

func NewConfirmDialog(
	title string,
	action string,
	callBack func(ctx app.Context, action bool)) *ConfirmDialog {
	return &ConfirmDialog{
		title:      title,
		actionText: action,
		callBack:   callBack,
	}
}

func (c *ConfirmDialog) Render() app.UI {
	return app.Div().Class("modal is-active").
		Body(
			app.Div().Class("modal-background"),
			app.Div().Class("modal-card").Body(
				app.Header().Class("modal-card-head").Body(
					app.P().Class("modal-card-title").Text(c.title),
					app.Button().Class("delete").Aria("label", "close").OnClick(c.onClose),
				),
				app.Footer().Class("modal-card-foot").Body(
					app.Button().Class("button is-success").Text(c.actionText).OnClick(c.onConfirm),
					app.Button().Text("Annuler").OnClick(c.onClose),
				),
			),
		)

}

func (c *ConfirmDialog) onClose(ctx app.Context, e app.Event) {
	c.callBack(ctx, false)
}
func (c *ConfirmDialog) onConfirm(ctx app.Context, e app.Event) {
	c.callBack(ctx, true)
}
