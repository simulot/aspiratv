package bulma

import (
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type TextField struct {
	app.Compo
	Label        string
	Placeholder  string
	Value        *string
	HelpClass    string
	Help         string
	CheckHandler func(app.Context, string) (string, error)
}

func NewTextField(value *string, label string, placeholder string) *TextField {
	return &TextField{
		Label:       label,
		Placeholder: placeholder,
		Value:       value,
	}
}

func (t *TextField) WithCheckHandle(handler func(ctx app.Context, value string) (string, error)) *TextField {
	t.CheckHandler = handler
	return t
}

func (t *TextField) Render() app.UI {
	return app.Div().
		Class("field").
		Body(
			app.Label().
				Class("label").
				Text(t.Label),
			app.Div().
				Class("control").
				Body(
					app.Input().
						Class("input").
						Type("text").
						Placeholder(t.Placeholder).
						Value(*t.Value).
						OnInput(t.onInput),
				),
			app.P().Class(t.HelpClass).Text(t.Help),
		)
}

func (t *TextField) onInput(ctx app.Context, e app.Event) {
	*t.Value = ctx.JSSrc.Get("value").String()
	if t.CheckHandler != nil {
		help, err := t.CheckHandler(ctx, *t.Value)
		if len(help) > 0 {
			t.Help = help
			if err != nil {
				t.HelpClass = "is-error"
			} else {
				t.HelpClass = "is-success"
			}
		}
	}
}
