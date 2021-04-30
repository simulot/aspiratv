package bulma

import (
	"fmt"
	"strings"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type Notification struct {
	app.Compo
	id        string
	class     []string
	text      string
	onDelete  func()
	hasDelete bool
}

func NewNotification() *Notification {
	return &Notification{}
}

func (n *Notification) Text(text string) *Notification {
	n.text = text
	return n
}

func (n *Notification) Class(c string) *Notification {
	n.class = append(n.class, c)
	return n
}

func (n *Notification) Delete(fn func()) *Notification {
	n.onDelete = fn
	n.hasDelete = true
	return n
}

func (n *Notification) Render() app.UI {
	return app.Div().
		Class("notification").
		Class(strings.Join(n.class, " ")).
		Body(
			app.Raw(fmt.Sprintf("<!-- %#v -->", n)),
			app.Text(n.text),
			app.If(n.hasDelete,
				app.Button().
					Class("delete").
					OnClick(func(ctx app.Context, e app.Event) {
						n.onDelete()
					}, n),
			),
		)
}
