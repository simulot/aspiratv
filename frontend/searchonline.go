package frontend

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/maxence-charriere/go-app/v8/pkg/app"
	"github.com/simulot/aspiratv/models"
)

const (
	tagAll        = "Afficher tout"
	tagVisible    = "Visibles"
	tagCollection = "Collections"
	tagSeries     = "Series"
	tagMedia      = "Medias"
	tagMagazine   = "Magazines"
)

type tagBadge struct {
	Selected bool
	Count    int
}

type resultItem struct {
	models.SearchResult
	tags []string
}

type SearchOnline struct {
	app.Compo

	search       string
	Results      []resultItem
	IsRunning    bool
	cancelChanel chan struct{}
	num          int
	Tags         map[string]tagBadge
	All          tagBadge
}

func (c *SearchOnline) renderTag(k string) app.UI {
	var tag tagBadge
	if k == tagAll {
		tag = c.All
	} else {
		tag = c.Tags[k]
	}
	return app.Div().Class("control").Body(
		app.Span().Class("tag").Class(ifThenElseString(tag.Selected, "is-info", "is-light")).Body(
			app.Text(k+" ("),
			app.Text(strconv.Itoa(tag.Count)),
			app.Text(") "),
			app.Button().Class("delete is-small"),
		),
	).OnClick(c.OnTagClick(k))
}

func (c *SearchOnline) OnTagClick(t string) func(ctx app.Context, e app.Event) {
	return func(ctx app.Context, e app.Event) {
		if t == tagAll {
			c.All.Selected = !c.All.Selected
			if c.All.Selected {
				for k, t := range c.Tags {
					t.Selected = true
					c.Tags[k] = t
				}
			}

		} else {
			tag := c.Tags[t]
			tag.Selected = !tag.Selected
			c.Tags[t] = tag
		}
		c.All.Selected = true
		for _, t := range c.Tags {
			if !t.Selected {
				c.All.Selected = false
			}
		}

		c.Update()
	}
}

func (c *SearchOnline) isResultVisible(r resultItem) bool {
	for _, t := range r.tags {
		if !c.Tags[t].Selected {
			return false
		}
	}
	return true
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
			app.Div().Class("field is-grouped is-grouped-multiline").Body(
				c.renderTag(tagAll),
				app.Range(c.Tags).Map(func(k string) app.UI {
					return c.renderTag(k)
				}),
			),
			app.If(len(c.Results) > 0,
				app.Div().Body(
					app.H1().Class("subtitle").Text(fmt.Sprintf("%d résultat(s)", len(c.Results))),
					app.If(c.IsRunning, app.Button().Text("Arrêter").OnClick(c.ClickOnCancel)),
				),
				app.Div().Class("columns is-multiline is-mobile").Body(
					app.Range(c.Results).Slice(func(i int) app.UI {
						r := c.Results[i]
						if !c.isResultVisible(r) {
							return nil
						}
						return app.Div().Class("column is-4").Body(
							app.Div().Class("card").Body(
								app.Div().Class("card-image").Body(app.Img().Class("image").Src(r.ThumbURL)),
								app.Div().Class("card-header").Body(
									app.Div().Class("card-header-title").Text(r.Title),
								),
								app.Div().Class("card-content").Body(
									app.Div().Class("tags").Body(
										app.Span().Class("tag").Text(r.Chanel),
										app.Span().Class("tag").Body(
											app.If(r.Type == models.MediaTypeCollection, app.Text("Collection")).
												ElseIf(r.Type == models.MediaTypeSeries, app.Text("Série")).
												ElseIf(r.Type == models.MediaTypeTVShow, app.Text("Magazine")).
												ElseIf(r.Type == models.MediaTypeMovie, app.Text("Media")),
										),
										app.If(r.IsTeaser, app.Span().Class("tag").Text("Bande annonce")),
										app.If(r.Type != models.MediaTypeCollection && r.Type != models.MediaTypeSeries,
											app.If(r.AvailableOn.IsZero() || r.AvailableOn.Before(time.Now()), app.Span().Class("tag").Text("Visible")).
												Else(app.Span().Class("tag").Text(r.AvailableOn.Local().Format("Disponible le 02/01 à 15h04"))),
										),
									),
									app.Div().Class("content").Body(app.Text(r.Plot)),
								),
							),
						)
					}),
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

	c.Results = []resultItem{}
	c.All = tagBadge{Selected: true}
	c.Tags = map[string]tagBadge{}
	c.num++

	c.IsRunning = true
	c.cancelChanel = make(chan struct{})

	q := models.SearchQuery{
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
				result := resultItem{
					SearchResult: r,
				}

				all := c.All
				all.Count++
				c.All = all

				c.addTag(&result, r.Chanel)
				if r.IsPlayable {
					c.addTag(&result, tagVisible)
				}
				switch r.Type {
				case models.MediaTypeCollection:
					c.addTag(&result, tagCollection)
				case models.MediaTypeSeries:
					c.addTag(&result, tagSeries)
				case models.MediaTypeTVShow:
					c.addTag(&result, tagMagazine)
				case models.MediaTypeMovie:
					c.addTag(&result, tagMedia)
				}

				c.Results = append(c.Results, result)
				c.Update()
			}
		}
	})
}

func (c *SearchOnline) addTag(r *resultItem, tag string) {
	r.tags = append(r.tags, tag)
	t := c.Tags[tag]
	t.Count++
	t.Selected = true
	c.Tags[tag] = t
}

func ifString(c bool, s string) string {
	if !c {
		return ""
	}
	return s
}

func ifThenElseString(c bool, whenTrue string, whenFalse string) string {
	if !c {
		return whenFalse
	}
	return whenTrue
}
