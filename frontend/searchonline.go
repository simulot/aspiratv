package frontend

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/maxence-charriere/go-app/v8/pkg/app"
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

	IsRunning     bool
	ProvidersTags *ProviderTags
	ResultCards   *ResultCards
	ChannelsList  *ChanneList

	search string
	strict bool
}

func (c *SearchOnline) Render() app.UI {
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
			app.If(c.ResultCards != nil, c.ResultCards),
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

	if c.ProvidersTags == nil {
		return
	}

	cancellableCtx, cancel := context.WithCancel(ctx)
	if c.IsRunning {
		cancel()
		c.IsRunning = false
		ctx.Dispatch(c.Update)
		return
	}

	c.IsRunning = true
	c.ResultCards = NewResultCards(c.ChannelsList)

	// Wait click on Cancel button
	go func() {
		<-cancellableCtx.Done()
		c.IsRunning = false
	}()

	q := models.SearchQuery{
		Title:          c.search,
		OnlyExactTitle: c.strict,
		//TODO add selected channels
	}
	results, err := MyAppState.s.Search(cancellableCtx, q)
	if err != nil {
		log.Printf("Search API returns error: %s", err)
		cancel()
		return
	}

	// Wait results delivered by the server
	ctx.Async(func() {
		defer func() {
			c.IsRunning = false
			cancel()
			c.Update()
		}()
		for {
			select {
			case <-cancellableCtx.Done():
				return
			case r, ok := <-results:
				if !ok {
					return
				}
				c.ResultCards.AddResult(r)
			}
		}
	})

	c.Update()
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
		c.ResultCards = NewResultCards(c.ChannelsList)

		c.ProvidersTags = NewProviderTags()

		for _, ch := range c.ChannelsList.SortedList() {
			c.ProvidersTags.Tags.SetTag(&TagInfo{Code: ch.Code, State: TagSelected, Text: ch.Name})
		}
		ctx.Dispatch(func() {
			c.Update()
			log.Print("Providers loaded and updated")
		})
	})
}

type ProviderTags struct {
	app.Compo
	Tags *TagList
}

func NewProviderTags() *ProviderTags {
	return &ProviderTags{
		Tags: NewTagList(&TagListOptions{CanDisable: false, All: &TagInfo{Icon: app.I().Class("mdi mdi-television-classic is-medium"), State: TagSelected}}),
	}
}

func (pt *ProviderTags) Render() app.UI {
	if pt.Tags == nil {
		return app.Text("Loading channel list...")
	}
	return pt.Tags
}

type ResultCards struct {
	app.Compo
	Cards      []*Card
	Channels   *TagList
	Tags       *TagList
	ChanneList *ChanneList
	MediaTypes *TagList
}

func NewResultCards(channels *ChanneList) *ResultCards {
	c := ResultCards{
		ChanneList: channels,
	}
	c.Reset()
	return &c
}

func (c *ResultCards) OnMount(ctx app.Context) {
	log.Printf("ResultCards mounted")
}

func (c *ResultCards) Reset() {
	c.Cards = []*Card{}
	c.Channels = NewTagList(&TagListOptions{CanCount: true, All: &TagInfo{State: TagSelected, Icon: app.I().Class("mdi mdi-television-classic")}})
	c.Tags = NewTagList(&TagListOptions{CanCount: true, All: &TagInfo{State: TagSelected, Icon: app.I().Class("mdi mdi-tag-multiple")}})
	c.MediaTypes = NewTagList(&TagListOptions{CanCount: true})
}

func (c *ResultCards) Render() app.UI {
	log.Printf("Len MediaTypes %d", len(c.MediaTypes.tags))
	return app.Div().Class("box").Body(
		app.P().Body(c.MediaTypes),
		app.P().Body(c.Channels),
		app.P().Body(c.Tags),
		app.Div().Class("columns is-multiline is-mobile").Body(
			app.Range(c.Cards).Slice(func(i int) app.UI {
				if c.IsCardVisible(c.Cards[i]) {
					return c.Cards[i]
				}
				return nil
			}),
		),
	)
}

func (c *ResultCards) IsCardVisible(card *Card) bool {
	if c.Channels.GetState(card.Chanel) != TagSelected {
		return false
	}

	return true
}

func (c *ResultCards) AddResult(r models.SearchResult) {
	card := NewCard(c.ChanneList, r)
	c.Tags.IncAll()
	for _, t := range r.Tags {
		c.Tags.SetOrIncTag(&TagInfo{Code: t, Text: t, State: TagSelected})
	}
	c.Channels.IncAll()
	c.Channels.SetOrIncTag(&TagInfo{Code: r.Chanel, State: TagSelected})

	c.MediaTypes.SetTag(&TagInfo{Code: r.Type.String(), Text: models.MediaTypeLabel[r.Type]})

	c.Cards = append(c.Cards, card)
	// sort.Slice(c.Cards, func(i, j int) bool {
	// 	return c.Cards[i].Title < c.Cards[j].Title
	// })
	if c.Mounted() {
		c.Update()
	} else {
		log.Printf("Results not mounted")
	}
}

type Card struct {
	app.Compo
	models.SearchResult
	Tags       *TagList
	ChanneList *ChanneList
}

func NewCard(ChanneList *ChanneList, r models.SearchResult) *Card {
	c := Card{
		Tags:         NewTagList(nil),
		SearchResult: r,
		ChanneList:   ChanneList,
	}

	c.Tags.SetTag(&TagInfo{Code: r.Type.String(), Text: models.MediaTypeLabel[r.Type]})
	for _, t := range r.Tags {
		c.Tags.SetTag(&TagInfo{Code: t, Text: t})
	}
	return &c
}

func (c *Card) Render() app.UI {
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
