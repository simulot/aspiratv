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
)

type PageID int

const (
	PageSearchOnLine PageID = iota
	PageLibrary
	PageSettings
	PageCredits
)

// AppState hold the state of the application
type AppState struct {
	Ready chan struct{}

	// When running in backend
	InBackend bool

	// Indicate when the application is ready. Used by LoadSetting component
	StateReady bool

	// API use server's REST API
	API *API

	// Application settings
	Settings models.Settings

	// Keep track of displayed page, used for application menue
	CurrentPage PageID

	// List of menu items.
	menuItems []Menuitem

	// Dispatch sent notifications to all of its subscribers
	Dispatch *dispatcher.Dispatcher

	// Drawer display notifications
	Drawer *NotificationsDrawer

	// List of available channels and TV sites
	ChannelsList *ChanneList

	// For Search Page
	// Store results and presents them back instantly
	Results []models.SearchResult

	serverNotificationCancel func()
	StateContext             context.Context
}

var MyAppState *AppState

func InitializeWebApp(ctx context.Context) *AppState {
	u := app.Window().URL()
	u.Scheme = "http"
	u.Path = "/api/"
	s := NewRestStore(u.String())
	MyAppState = NewAppState(ctx, s)

	go func() {
		ps, err := MyAppState.API.ProviderDescribe(ctx)
		if err != nil {
			log.Print("Providers error: ", err)
			return
		}
		MyAppState.ChannelsList = NewChannelList(ps)
		MyAppState.Settings, err = MyAppState.GetSettings(ctx)
		if err != nil {
			log.Print("Settings error: ", err)
			return
		}

		MyAppState.StateReady = true
		close(MyAppState.Ready)
	}()
	return MyAppState
}

func NewAppState(ctx context.Context, s *API) *AppState {

	state := AppState{
		StateContext: ctx,
		Ready:        make(chan struct{}, 1),
		API:          s,
		CurrentPage:  PageSearchOnLine,
		menuItems: []Menuitem{
			{
				PageSearchOnLine,
				"",
				"Chercher en ligne",
				"/search",
			},
			{
				PageLibrary,
				"",
				"Bibliothèque",
				"/library",
			},
			{
				PageSettings,
				"",
				"Réglages",
				"/settings",
			},
			{
				PageCredits,
				"",
				"Crédits",
				"/credits",
			},
		},
	}
	state.Dispatch = dispatcher.NewDispatcher()
	state.Drawer = NewNotificationsDrawer()
	state.Drawer.Attach(state.Dispatch)
	// if !app.IsServer {
	// 	go state.ServerNotifications(ctx)
	// }
	return &state
}

func (s *AppState) GetSettings(ctx context.Context) (models.Settings, error) {
	if !app.IsServer {
		settings, err := MyAppState.API.GetSettings(ctx)
		if err != nil {
			return models.Settings{}, err
		}
		return settings, nil
	}
	return models.Settings{}, nil
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

	passage := 0
	go func() {
		connected := false
		errMessage := models.NewMessage("connexion", models.StatusSuccess)
		for {
			log.Printf("[NOTIFICATION CLIENT] Connecting to notification server #%d", passage)
			passage++
			messages, err := s.API.SubscribeServerNotifications(ctx)

			select {
			case <-ctx.Done():
				return
			default:

				if err != nil {
					log.Printf("[NOTIFICATION CLIENT] Can't get connection with the server")
					if connected {
						errMessage.Text = "Connexion perdue avec le serveur"
					} else {
						errMessage.Text = "Connexion impossible avec le serveur"
					}
					errMessage.Status = models.StatusError
					s.Dispatch.Publish(errMessage)
					time.Sleep(5 * time.Second)
					continue
				} else {
					if errMessage.Status == models.StatusError {
						connected = true
						errMessage.Text = "Connexion rétablie avec le serveur"
						errMessage.Status = models.StatusSuccess
						s.Dispatch.Publish(errMessage)

					}
				}
			}
			log.Printf("[NOTIFICATION CLIENT] Connected to notification server")

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
