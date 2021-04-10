package frontend

import (
	"context"
	"fmt"
	"log"

	"github.com/maxence-charriere/go-app/v8/pkg/app"
	"github.com/simulot/aspiratv/store"
)

type SearchOnline struct {
	app.Compo

	search       string
	Results      []string
	IsRunning    bool
	cancelChanel chan struct{}
	num          int
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
					app.If(c.IsRunning, app.Button().Text("Arrêter").OnClick(c.ClickOnCancel)),
					// app.Button().Text("Effacer").OnClick()
					app.Div().Class("columns is-multiline").Body(
						app.Range(c.Results).Slice(func(i int) app.UI {
							return app.Div().Class("column is-6 box").Body(
								app.P().Class("title").Text(c.Results[i]),
								app.P().Text("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Etiam semper diam at erat pulvinar, at pulvinar felis blandit. Vestibulum volutpat tellus diam, consequat gravida libero rhoncus ut. Morbi maximus, leo sit amet vehicula eleifend, nunc dui porta orci, quis semper odio felis ut quam."),
							)
						}),
					),
				),
			),
		),
	),
	)
}

func (c *SearchOnline) ClickOnCancel(ctx app.Context, e app.Event) {
	if c.IsRunning {
		close(c.cancelChanel)
	}
}

func (c *SearchOnline) ClickOnSearch(ctx app.Context, e app.Event) {
	if c.IsRunning {
		return
	}
	c.Results = nil
	c.num++

	c.IsRunning = true
	c.cancelChanel = make(chan struct{})

	q := store.SearchQuery{
		Title: c.search,
	}

	cancellableCtx, cancel := context.WithCancel(ctx)

	go func() {
		// Wait click on Cancel button
		select {
		case <-c.cancelChanel:
			log.Print("Cancel button hit")
			cancel()
			return
		case <-cancellableCtx.Done():
			return
		}
	}()

	results, err := MyAppState.s.Search(cancellableCtx, q)
	if err != nil {
		log.Printf("Search API returns error: %s", err)
		c.Results = []string{"Error " + err.Error()}
		return
	}

	c.Update()

	ctx.Async(func() {
		defer func() {
			c.IsRunning = false
			c.Update()
			cancel()
			log.Printf("Exit Async %d", c.num)
		}()
		for {
			select {
			case <-cancellableCtx.Done():
				return
			case r, ok := <-results:
				if !ok {
					return
				}
				c.Results = append(c.Results, r.Title)
				c.Update()
			}
		}
	})
}
