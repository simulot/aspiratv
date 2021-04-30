package frontend

import (
	"log"

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

// AppState hold the state of the application
type AppState struct {
	s           *RestClient
	currentPage PageID
	menuItems   []Menuitem

	Dispatch *models.NotificationDispatcher
	Drawer   *NotificationsDrawer
}

var MyAppState *AppState

func InitializeWebApp() *AppState {
	u := app.Window().URL()
	u.Scheme = "http"
	u.Path = "/api/"
	s := NewRestStore(u.String())
	MyAppState = NewAppState(s)

	return MyAppState
}

func NewAppState(s *RestClient) *AppState {
	dispatch := models.NewNotificationDispatcher()
	drawer := NewNotificationsDrawer()
	drawer.Attach(dispatch)
	return &AppState{
		s:           s,
		currentPage: PageSearchOnLine,
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
		Dispatch: dispatch,
		Drawer:   drawer,
	}
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
