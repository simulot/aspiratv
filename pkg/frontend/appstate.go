package frontend

import (
	"context"
	"log"
	"sort"
	"time"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/simulot/aspiratv/pkg/dispatcher"
	"github.com/simulot/aspiratv/pkg/models"
	"github.com/simulot/aspiratv/pkg/providers"
	"github.com/simulot/aspiratv/pkg/store"
)

//go:generate enumer -type=PageID -json
type PageID int

const (
	PageUndefined PageID = iota
	PageSearchOnLine
	PageLibrary
	PageSubscriptions
	PageEditSubscrition
	PageSettings
	PageCredits
)

// AppState hold the state of the application
type AppState struct {
	Ready chan struct{}

	// Indicate when the application is ready. Used by LoadSetting component
	StateReady bool

	// API use server's REST API
	API *APIClient

	Store store.Store

	// Application settings
	Settings models.Settings

	// Keep track of displayed page, used for application menue
	CurrentPage PageID

	// Dispatch sent notifications to all of its subscribers
	Dispatch *dispatcher.Dispatcher

	// Drawer display notifications
	Drawer *NotificationsDrawer

	// List of available channels and TV sites
	ChannelsList *ChanneList

	serverNotificationCancel func()
	StateContext             context.Context
}

var MyAppState *AppState

//InitializeWebApp initialize the client side either for the browser and the serverside rendering
func InitializeWebApp(ctx context.Context, endpoint string, st store.Store) *AppState {
	s := NewAPIClient(endpoint, st)
	MyAppState = NewAppState(ctx, s)

	// go
	func() {
		log.Printf("InitializeWebApp get settings")

		var err error
		MyAppState.Settings, err = MyAppState.GetSettings(ctx)
		if err != nil {
			log.Print("Settings error: ", err)
		}

		log.Printf("InitializeWebApp get providers")
		ps, err := MyAppState.API.ProviderDescribe(ctx)
		if err != nil {
			log.Print("Providers error: ", err)
		}
		MyAppState.ChannelsList = NewChannelList(ps)

		MyAppState.StateReady = true
		log.Printf("InitializeWebApp ready")
		close(MyAppState.Ready)
	}()
	return MyAppState
}

func NewAppState(ctx context.Context, s *APIClient) *AppState {

	state := AppState{
		StateContext: ctx,
		Ready:        make(chan struct{}, 1),
		API:          s,
		CurrentPage:  PageSearchOnLine,
	}
	state.Dispatch = dispatcher.NewDispatcher()
	state.Drawer = NewNotificationsDrawer()
	state.Drawer.Attach(state.Dispatch)
	if app.IsClient {
		go state.ServerNotifications(ctx)
	}
	return &state
}

func (s *AppState) GetSettings(ctx context.Context) (models.Settings, error) {
	settings, err := MyAppState.API.GetSettings(ctx)
	if err != nil {
		return models.Settings{}, err
	}
	return settings, nil
}

func (s *AppState) ToggleServerNotifications(b bool) {
	if b && s.serverNotificationCancel == nil {
		s.serverNotificationCancel = s.ServerNotifications(s.StateContext)
		return
	}
	if !b && s.serverNotificationCancel != nil {
		s.serverNotificationCancel()
		s.serverNotificationCancel = nil
	}
}

// Start ServerNotifications receiver and return a function to turning it off
func (s *AppState) ServerNotifications(ctx context.Context) func() {
	ctx, cancelCtx := context.WithCancel(ctx)
	closeFun := func() {
		cancelCtx()
	}

	go func() {
		connected := false
		errMessage := models.NewMessage("connexion").SetStatus(models.StatusSuccess)
		for {
			messages, err := s.API.SubscribeServerNotifications(ctx)

			select {
			case <-ctx.Done():
				return
			default:

				if err != nil {
					log.Printf("[NOTIFICATION CLIENT] Can't get connection with the server")
					if connected {
						errMessage.SetText("Connexion perdue avec le serveur")
					} else {
						errMessage.SetText("Connexion impossible avec le serveur")
					}
					errMessage.SetStatus(models.StatusError).SetPinned(true)
					s.Dispatch.Publish(errMessage)
					time.Sleep(5 * time.Second)
					continue
				} else {
					if errMessage.Status == models.StatusError {
						errMessage.SetText("Connexion rétablie avec le serveur").SetStatus(models.StatusSuccess).SetPinned(false)
						s.Dispatch.Publish(errMessage)
					}
				}
			}
			log.Printf("[NOTIFICATION CLIENT] Connected to notification server")
			connected = true
		messageLoop:
			for {
				select {
				case <-ctx.Done():
					return
				case m, ok := <-messages:
					if !ok {
						break messageLoop
					}
					s.Dispatch.Publish(m)
				}
			}
			log.Printf("[NOTIFICATION CLIENT] Lost connection with server")
		}

	}()

	return closeFun
}

func StringIf(b bool, whenTrue string, whenFalse string) string {
	if b {
		return whenTrue
	}
	return whenFalse
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

func Back(ctx app.Context) {
	ctx.NewAction("GotoBack").Post()
}

// GotoPage i
func GotoPage(ctx app.Context, pageID PageID, value interface{}) {
	// Query state of current page
	ctx.NewAction("BeforeMoving").Post()
	// Goto
	ctx.NewAction("GotoPage").Tag("page", pageID.String()).Value(value).Post()
}

type PageState struct {
	page  PageID      // Page ID to ease Goto
	title string      // Page name in clear for a potential bread crumb
	state interface{} // The state to set
}

func actionToState(action app.Action) PageState {
	page, _ := PageIDString(action.Tags.Get("page"))
	title := action.Tags.Get("title")
	return PageState{
		page:  page,
		title: title,
		state: action.Value,
	}
}
func stateToAction(s PageState) app.Action {
	a := app.Action{}
	a.Name = "Back"
	a.Tags.Set("page", s.page)
	a.Value = s.state
	return a
}
