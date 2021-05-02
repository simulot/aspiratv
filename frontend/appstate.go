package frontend

import (
	"context"
	"log"
	"time"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/simulot/aspiratv/models"
)

type PageID int

const (
	PageSearchOnLine PageID = iota
	PageLibrary
	PageSettings
	PageCredits
)

// sync.Condition:  https://kaviraj.me/understanding-condition-variable-in-go/

// AppState hold the state of the application
type AppState struct {
	Ready chan struct{}

	// Indicate when the application is ready. Used by LoadSetting component
	StateReady bool

	// Store use server's REST API
	Store *RestClient

	// Application settings
	Settings models.Settings

	// Keep track of displayed page, used for application menue
	CurrentPage PageID

	// List of menu items.
	menuItems []Menuitem

	// Dispatch sent notifications to all of its subscribers
	Dispatch *models.NotificationDispatcher

	// Drawer display notifications
	Drawer *NotificationsDrawer

	// List of available channels and TV sites
	ChannelsList *ChanneList

	// For Search Page
	// Store results and presents them back instantly
	Results []models.SearchResult
}

var MyAppState *AppState

func InitializeWebApp(ctx context.Context) *AppState {
	u := app.Window().URL()
	u.Scheme = "http"
	u.Path = "/api/"
	s := NewRestStore(u.String())
	MyAppState = NewAppState(ctx, s)

	go func() {
		ps, err := MyAppState.Store.ProviderDescribe(ctx)
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

		time.Sleep(2 * time.Second)
		MyAppState.StateReady = true
		close(MyAppState.Ready)
	}()
	return MyAppState
}

func NewAppState(ctx context.Context, s *RestClient) *AppState {

	state := AppState{
		Ready:       make(chan struct{}, 1),
		Store:       s,
		CurrentPage: PageSearchOnLine,
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
	state.Dispatch = models.NewNotificationDispatcher()
	state.Drawer = NewNotificationsDrawer()
	state.Drawer.Attach(state.Dispatch)

	return &state
}

func (s *AppState) GetSettings(ctx context.Context) (models.Settings, error) {
	settings, err := MyAppState.Store.GetSettings(ctx)
	if err != nil {
		return models.Settings{}, err
	}
	return settings, nil
}

func StringIf(b bool, whenTrue string, whenFalse string) string {
	if b {
		return whenTrue
	}
	return whenFalse
}

type NotificationsDrawer struct {
	n           []models.Notification
	subscribers []chan struct{}
}

func NewNotificationsDrawer() *NotificationsDrawer {
	d := NotificationsDrawer{}
	return &d
}

// Attach a notification provider to the drawer
func (d *NotificationsDrawer) Attach(sub models.Subscriber) {
	sub.Subscribe(d.onNotification)
}

func (d *NotificationsDrawer) onNotification(n models.Notification) {
	d.n = append(d.n, n)
	d.notify()
}

func (d *NotificationsDrawer) notify() {
	for _, c := range d.subscribers {
		c <- struct{}{}
	}
}

func (d *NotificationsDrawer) OnChange(fn func()) func() {
	c := make(chan struct{}, 1)
	go func() {
		for _ = range c {
			fn()
		}
	}()
	d.subscribers = append(d.subscribers, c)

	return func() {
		for i := 0; i < len(d.subscribers); i++ {
			if d.subscribers[i] == c {
				close(c)
				d.subscribers[i] = d.subscribers[len(d.subscribers)-1]
				d.subscribers[len(d.subscribers)-1] = nil
				d.subscribers = d.subscribers[0 : len(d.subscribers)-1]
			}
		}
	}
}

func (d *NotificationsDrawer) Notifications() []models.Notification {
	r := []models.Notification{}
	for i := 0; i < len(d.n); i++ {
		r = append(r, d.n[i])
	}
	return r
}

func (d *NotificationsDrawer) Dismiss(n models.Notification) {
	id := n.ID()
	for i := range d.n {
		if d.n[i].ID() == id {
			log.Printf("dismiss %d %s", i, d.n[i].Text)
			copy(d.n[i:], d.n[i+1:])                // Shift d.n[i+1:] left one index.
			d.n[len(d.n)-1] = models.Notification{} // Erase last element (write zero value).
			d.n = d.n[:len(d.n)-1]                  // Truncate slice.
			d.notify()
			return
		}
	}
}
