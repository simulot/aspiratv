package frontend

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/simulot/aspiratv/models"
	"github.com/simulot/aspiratv/providers"
)

const (
	tagAvailableOn       = "Disponible le 02/01 à 15h04"
	labelExactMatch      = "\u00a0Correspondance exacte du titre"
	labelOpenSource      = "Voir la page sur %s"
	labelAired           = "Diffusé le 02/01/2006 à 15h04"
	labelAvailableVideos = "%d%s video(s) disponible(s)"
)

type Search struct {
	app.Compo
	Results []models.SearchResult

	SearchTerms string
	IsRunning   bool
	StopSearch  chan struct{}
}

func (c *Search) OnMount(ctx app.Context) {
	log.Print("SearchOnLine mounted")
	MyAppState.CurrentPage = PageSearchOnLine
	c.Results = MyAppState.Results

	// ctx.Async(func() {
	// 	c.Update()

	// })
}

func (c *Search) OnDismount() {
	MyAppState.Results = c.Results
}

func (c *Search) OnUpdate(ctx app.Context) {
	log.Printf("SearchOnline OnUpdate")
}

func (c *Search) Render() app.UI {

	return AppPageRender(
		app.H1().Class("title is-1").Text("Rechercher sur l'Internet"),
		app.Div().Class("field is-groupped").Body(
			app.Div().Class("field has-addons").Body(
				app.Div().Class("control").Body(app.Input().Class("input").Type("text").Placeholder("keywords").AutoFocus(true).OnChange(c.ValueTo(&c.SearchTerms))),
				app.Div().Class("control").Body(app.Button().Class("button is-info").Text("Search").OnClick(c.Search).Class(StringIf(c.IsRunning, "is-loading", ""))),
			),
		),
		c.RenderResults(),
	)
}

func (c *Search) Search(ctx app.Context, e app.Event) {
	log.Printf("[SEARCH] %q", c.SearchTerms)

	if c.IsRunning {
		log.Printf("[SEARCH] Stopped clicked")
		close(c.StopSearch)
		c.IsRunning = false
		c.Update()
		return
	}

	c.StopSearch = make(chan struct{})
	c.IsRunning = true
	c.Results = []models.SearchResult{}
	go func() {
		cancelCtx, cancel := context.WithCancel(ctx)
		defer func() {
			log.Printf("[SEARCH] search goroutine ended")
			c.IsRunning = false
			cancel()
			c.Update()
		}()

		q := models.SearchQuery{
			Title: c.SearchTerms,
		}

		results, err := MyAppState.Store.Search(cancelCtx, q)
		if err != nil {
			log.Printf("[SEARCH] Search replies error: %s", err)
			return
		}

		for {
			select {
			case <-c.StopSearch:
				log.Printf("[SEARCH] Search stopped")
				return
			case <-cancelCtx.Done():
				log.Printf("[SEARCH] Cancellation: %s", cancelCtx.Err())
				close(c.StopSearch)
				return
			case r, ok := <-results:
				if !ok {
					log.Printf("[SEARCH] Search reply end")
					close(c.StopSearch)
					return
				}
				log.Printf("[SEARCH] Got %q", r.Title)
				c.Add(ctx, r)
			}
		}
	}()
	c.Update()

}

func (c *Search) Add(ctx app.Context, r models.SearchResult) {
	c.Results = append(c.Results, r)
	c.Update()
}

func (c *Search) RenderResults() app.UI {
	return app.If(len(c.Results) > 0,
		app.H2().Class("title is-2").Text(fmt.Sprintf("%d result(s)", len(c.Results))),
		app.Div().Class("columns is-multiline is-mobile").Body(
			app.Range(c.Results).Slice(func(i int) app.UI {
				return c.RenderResult(c.Results[i])
			}),
		),
	).Else(
		app.H2().Class("title is-2").Text("No result"),
	)
}

func (c *Search) RenderResult(r models.SearchResult) app.UI {
	log.Printf("SearchOnline rendering RenderResult %s", r.Title)
	return app.Body().Class("column is-6").Body(
		app.Div().Class("card").Body(
			app.Div().Class("card-image").Body(app.Img().Class("image").Src(r.ThumbURL)),
			app.Div().Class("card-content").Body(
				app.Div().Class("media").Body(
					app.Div().Class("media-left").Body(app.Img().Class("image is-48x48").Src(MyAppState.ChannelsList.Channel(r.Chanel).Logo)).Title(MyAppState.ChannelsList.Channel(r.Chanel).Name),
					app.Div().Class("media-content").Body(
						app.P().Class("title is-6").Text(r.Show),
						app.P().Class("subtitle is-6").Text(r.ID),
						app.P().Class("subtitle is-6").Text(models.MediaTypeLabel[r.Type]),
						app.P().Class("subtitle is-6").Text(fmt.Sprintf(labelAvailableVideos, r.AvailableVideos, StringIf(r.MoreAvailable, "+", ""))),
						// app.P().Body(c.Tags.Render()),
					),
				),
				app.Div().Class("plot").Body(
					app.P().Text(r.Plot),
					// app.Raw(c.Plot),
				),
			),
			app.Div().Class("card-footer").Body(
				app.A().Href("#").Class("card-footer-item").Text("Télécharger"),
				app.A().Href("#").Class("card-footer-item").Text("Surveiller"),
			),
		),
	)

}

type ChanneList struct {
	channels map[string]providers.Channel
}

func NewChannelList(l []providers.Description) *ChanneList {
	c := ChanneList{
		channels: map[string]providers.Channel{},
	}

	for _, p := range l {
		for code, ch := range p.Channels {
			c.channels[code] = ch
		}
	}
	return &c
}

func (c ChanneList) SortedList() []providers.Channel {
	s := []providers.Channel{}
	for _, ch := range c.channels {
		s = append(s, ch)
	}
	sort.Slice(s, func(i, j int) bool {
		return s[i].Name < s[j].Name
	})
	return s
}

func (c ChanneList) Channel(code string) providers.Channel { return c.channels[code] }

type DownloadDialog struct {
	Result models.SearchResult

	Path string
}

func NewDownloadDialog(r models.SearchResult) *DownloadDialog {
	return &DownloadDialog{
		Result: r,
	}
}
