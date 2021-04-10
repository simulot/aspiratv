package frontend

import (
	"fmt"
	"log"

	"github.com/maxence-charriere/go-app/v8/pkg/app"
)

type SearchOnline struct {
	app.Compo

	search  string
	Results []string
}

func (c *SearchOnline) Render() app.UI {
	MyAppState.currentPage = PageSearchOnLine
	return app.Div().Class("container").Body(app.Div().Class("columns").Body(
		&MyApp{},
		app.Div().Class("column").Body(
			app.H2().Class("title").Text("Rechercher en ligne"),
			app.Div().Class("field has-addons").Body(
				app.Div().Class("control").Body(app.Input().Class("input").Type("text").Placeholder("Chercher une émission ou série en ligne").AutoFocus(true).OnChange(c.ValueTo(&c.search))),
				app.Div().Class("control").Body(app.Button().Class("button is-info").Text("Chercher").OnClick(c.ClickOnSearch)),
			),
			app.If(len(c.Results) > 0,
				app.Section().Class("section").Body(
					app.H1().Class("subtitle").Text(fmt.Sprintf("%d résultat(s)", len(c.Results))),
					// app.Button().Text("Effacer").OnClick()
					app.Range(c.Results).Slice(func(i int) app.UI {
						return app.Div().Body(
							app.Text(c.Results[i]),
							app.Br(),
						)
					}),
				),
			),
		),
	),
	)
}

func (c *SearchOnline) ClickOnSearch(ctx app.Context, e app.Event) {
	c.Results = nil
	results, err := MyAppState.s.Search(ctx)
	if err != nil {
		log.Printf("Search API returns error: %s", err)
		c.Results = []string{"Error " + err.Error()}
		return
	}

	ctx.Async(func() {
		for r := range results {
			c.Results = append(c.Results, r.Title)
			c.Update()
		}
	})
}
