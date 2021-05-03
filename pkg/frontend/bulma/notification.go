package bulma

import (
	"strings"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type Notification struct {
	app.Compo
	Cl             []string
	T              string
	DeleteCallback func()
}

func NewNotification() *Notification {
	return &Notification{
		DeleteCallback: func() {},
	}
}

func (n *Notification) Text(text string) *Notification {
	n.T = text
	return n
}

func (n *Notification) Class(c string) *Notification {
	n.Cl = append(n.Cl, c)
	return n
}

func (n *Notification) OnDelete(fn func()) *Notification {
	n.DeleteCallback = fn
	return n
}

func (n *Notification) Render() app.UI {
	return app.Div().
		Class("notification").
		Class(strings.Join(n.Cl, " ")).
		Body(
			app.Text(n.T),
			app.If(n.DeleteCallback != nil,
				app.Button().
					Class("delete").
					OnClick(func(ctx app.Context, e app.Event) {
						n.DeleteCallback()
					}, n),
			),
		)
}
