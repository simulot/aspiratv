package frontend

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/simulot/aspiratv/models"
	"github.com/simulot/aspiratv/providers"
)

const (
	tagAvailableOn  = "Disponible le 02/01 à 15h04"
	labelExactMatch = "\u00a0Correspondance exacte du titre"
	labelOpenSource = "Voir la page sur %s"
	labelAired      = "Diffusé le 02/01 à 15h04"
)

// SearchOnline is one of root pages. Can't be initialized statically
type SearchOnline struct {
	app.Compo

	ProvidersTags *ProviderTags
	ResultCards   *ResultCards
	ChannelsList  *ChanneList

	search    string
	strict    bool
	IsRunning bool
	stop      chan struct{}
}

func (c *SearchOnline) Render() app.UI {
	log.Printf("SearchOnline.Render")
	if c.ResultCards != nil && c.ResultCards.Cards != nil {
		log.Printf("--SearchOnline.Render.ResultCards %d", len(c.ResultCards.Cards))
	}
	MyAppState.currentPage = PageSearchOnLine
	return app.Div().Class("container").Body(app.Div().Class("columns").Body(
		&MyApp{},
		app.Div().Class("column").Body(
			app.H2().Class("title").Text("Rechercher en ligne"),
			app.Div().Class("field is-groupped").Body(
				app.Div().Class("field has-addons").Body(
					app.Div().Class("control").Body(app.Input().Class("input").Type("text").Placeholder("Chercher une émission ou série en ligne").AutoFocus(true).OnChange(c.ValueTo(&c.search))),
					app.Div().Class("control").Body(app.Button().Class("button is-info").Class(StringIf(c.IsRunning, "is-loading", "")).Text("Chercher").OnClick(c.ClickOnSearch)),
				),
				app.Div().Class("control").Body(app.Label().Class("checkbox").Body(app.Input().Type("checkbox").OnChange(c.onTick(&c.strict)), app.Text(labelExactMatch))),
				c.ProvidersTags,
			),
			c.ResultCards,
			// app.If(c.ResultCards != nil), // app.P().Text(fmt.Sprintf("%d réponses.", len(c.ResultCards.Cards))),
			// 	// c.ResultCards,

		),
	),
	)
}

func (c *SearchOnline) onTick(b *bool) func(ctx app.Context, e app.Event) {
	return func(ctx app.Context, e app.Event) {
		v := ctx.JSSrc.Get("value").String()
		*b = v == "on"
	}
}

func (c *SearchOnline) ClickOnSearch(ctx app.Context, e app.Event) {
	log.Printf("ClickOnSearch")
	defer c.Update()

	if c.IsRunning {
		log.Printf("ClickOnSearch: STOP")
		close(c.stop)
		return
	}

	c.IsRunning = true
	c.stop = make(chan struct{})
	c.ResultCards = NewResultCards(c.ChannelsList)

	go func() {
		cancelCtx, cancel := context.WithCancel(ctx)
		cleanExit := "Crash!"

		defer func() {
			log.Printf("ClickOnSearch goroutine ended Exit: %v", cleanExit)
			c.IsRunning = false
			cancel()
			c.Update()
		}()

		q := models.SearchQuery{
			Title:          c.search,
			OnlyExactTitle: c.strict,
			//TODO add selected channels
		}

		results, err := MyAppState.s.Search(cancelCtx, q)
		if err != nil {
			log.Printf("Search API returns error: %s", err)
			cancel()
			close(c.stop)
			return
		}

		for {
			select {
			case <-c.stop:
				cleanExit = "Stop detected"
				return
			case <-cancelCtx.Done():
				close(c.stop)
				cleanExit = "<-cancelCtx.Done()"
				return
			case r, ok := <-results:
				if !ok {
					close(c.stop)
					cleanExit = "End of results channel"
					return
				}
				// Touch the UI in the Dispatch context
				ctx.Dispatch(func(app.Context) {
					c.ResultCards.AddResult(r)
					c.Update()
				})
			}
		}
	}()

}

func firstStringOf(ss ...string) string {
	for _, s := range ss {
		if len(s) > 0 {
			return s
		}
	}
	return ""
}

func (c *SearchOnline) OnMount(ctx app.Context) {
	log.Print("SearchOnLine mounted")

	ctx.Async(func() {
		ps, err := MyAppState.s.ProviderDescribe(ctx)
		if err != nil {
			log.Print("Providers error: ", err)
			return
		}
		c.ChannelsList = NewChannelList(ps)
		c.ProvidersTags = NewProviderTags()
		for _, ch := range c.ChannelsList.SortedList() {
			c.ProvidersTags.SetTag(&TagInfo{Code: ch.Code, State: TagSelected, Text: ch.Name})
		}
		c.ResultCards = NewResultCards(c.ChannelsList)
		c.Update()
	})
}

type ProviderTags struct {
	app.Compo
	*TagList
}

func NewProviderTags() *ProviderTags {
	return &ProviderTags{
		TagList: NewTagList(&TagListOptions{CanDisable: false, All: &TagInfo{Icon: app.I().Class("mdi mdi-television-classic is-medium"), State: TagSelected}}),
	}
}

func (pt *ProviderTags) Render() app.UI {
	log.Printf("Rendering ProviderTags")

	if len(pt.TagList.tags) == 0 {
		return app.Text("Loading channel list...")
	}
	return pt.TagList
}

type ResultCards struct {
	app.Compo
	Cards      []*Card
	ChanneList *ChanneList
	// Channels   *TagList
	// Tags       *TagList
	// MediaTypes *TagList
}

func NewResultCards(channels *ChanneList) *ResultCards {
	c := ResultCards{
		ChanneList: channels,
	}
	c.initialize()
	return &c
}

func (c *ResultCards) OnMount(ctx app.Context) {
	log.Printf("ResultCards mounted")
}

func (c *ResultCards) initialize() {
	c.Cards = []*Card{}
	// c.Channels = NewTagList(&TagListOptions{CanCount: true, All: &TagInfo{State: TagSelected, Icon: app.I().Class("mdi mdi-television-classic")}})
	// c.Tags = NewTagList(&TagListOptions{CanCount: true, All: &TagInfo{State: TagSelected, Icon: app.I().Class("mdi mdi-tag-multiple")}})
	// c.MediaTypes = NewTagList(&TagListOptions{CanCount: true})
}

func (c *ResultCards) Render() app.UI {
	log.Printf("ResultCards.Render, Cards %d", len(c.Cards))
	return app.Div().Class("box").Body(
		// app.P().Body(c.MediaTypes),
		// app.P().Body(c.Channels),
		// app.P().Body(c.Tags),
		app.Div().Class("columns is-multiline is-mobile").Body(
			app.Range(c.Cards).Slice(func(i int) app.UI {
				return c.Cards[i]

				// if c.IsCardVisible(c.Cards[i]) {
				// 	return c.Cards[i]
				// }
				// return nil
			}),
		),
	)
}

func (c *ResultCards) IsCardVisible(card *Card) bool {
	// if c.Channels.GetState(card.Chanel) != TagSelected {
	// 	return false
	// }

	return true
}

func (c *ResultCards) AddResult(r models.SearchResult) {
	log.Printf("AddResult %#v", r.ID)
	card := NewCard(c.ChanneList, r)
	// c.Tags.IncAll()
	// for _, t := range r.Tags {
	// 	c.Tags.SetOrIncTag(&TagInfo{Code: t, Text: t, State: TagSelected})
	// }
	// c.Channels.IncAll()
	// c.Channels.SetOrIncTag(&TagInfo{Code: r.Chanel, State: TagSelected})

	// c.MediaTypes.SetTag(&TagInfo{Code: r.Type.String(), Text: models.MediaTypeLabel[r.Type]})

	c.Cards = append(c.Cards, card)
	// sort.Slice(c.Cards, func(i, j int) bool {
	// 	return c.Cards[i].Title < c.Cards[j].Title
	// })
	// c.Defer(func(app.Context) {
	// 	c.Update()
	// })

}

type Card struct {
	app.Compo
	models.SearchResult
	Tags       *TagList
	ChanneList *ChanneList
}

func NewCard(ChanneList *ChanneList, r models.SearchResult) *Card {
	c := Card{
		// Tags:         NewTagList(nil),
		SearchResult: r,
		ChanneList:   ChanneList,
	}

	// c.Tags.SetTag(&TagInfo{Code: r.Type.String(), Text: models.MediaTypeLabel[r.Type]})
	// for _, t := range r.Tags {
	// 	c.Tags.SetTag(&TagInfo{Code: t, Text: t})
	// }
	return &c
}

func (c *Card) Render() app.UI {
	log.Printf("Rendering Card %#v", c.ID)
	return app.Div().Class("column is-4").Body(
		app.Div().Class("card").Body(
			app.Div().Class("card-image").Body(app.Img().Class("image").Src(c.ThumbURL)),
			app.Div().Class("card-content").Body(
				app.Div().Class("media").Body(
					app.Div().Class("media-left").Body(app.Img().Class("image is-48x48").Src(c.ChanneList.Channel(c.Chanel).Logo)).Title(c.ChanneList.Channel(c.Chanel).Name),
					app.Div().Class("media-content").Body(
						app.P().Class("title is-6").Text(firstStringOf(c.Show, c.Title)),
						app.If(c.Episode != 0 || c.Season != 0,
							app.P().Class("subtitle is-6").Body(
								app.If(c.Season != 0, app.Text(fmt.Sprintf("S%d", c.Season))),
								app.If(c.Episode != 0, app.Text(fmt.Sprintf("E%02d ", c.Episode))),
								app.Text(c.Title),
							),
						),
						app.If(!c.Aired.IsZero(), app.P().Class("subtitle is-6").Text(c.Aired.Format(labelAired))),
						app.P().Class("subtitle is-6").Text(c.ID),
						app.If(time.Now().Before(c.AvailableOn),
							app.P().Class("subtitle is-6").Text(c.AvailableOn.Local().Format(tagAvailableOn)),
						),
					),
				),
				app.Div().Class("content").Body(
					app.P().Body(c.Tags),
					app.P().Text(c.Plot),
					// app.Raw(c.Plot),
				),
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
