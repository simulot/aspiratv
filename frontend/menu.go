package frontend

import (
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type Menuitem struct {
	page  PageID
	icon  string
	label string
	path  string
}

type Menu struct {
	app.Compo
}

func (c *Menu) Render() app.UI {
	return app.Aside().Class("menu").Body(
		app.Ul().Class("menu-list").Body(
			app.Range(MyAppState.menuItems).Slice(func(i int) app.UI {
				return app.Li().Body(
					app.A().Class(StringIf(MyAppState.menuItems[i].page == MyAppState.currentPage, "is-active", "")).Href(MyAppState.menuItems[i].path).Text(MyAppState.menuItems[i].label))
			}),
		),
	)
}

func AppPageRender(pages ...app.UI) app.UI {
	return app.Div().Class("container").Body(app.Div().Class("columns").Body(
		&MyApp{},
		app.Div().Class("column").Body(
			app.Range(pages).Slice(func(i int) app.UI {
				return pages[i]
			}),
		),
	),
	)
}
