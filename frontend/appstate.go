package frontend

import (
	"log"
	"sync"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
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

	Messages *MessageDispatcher
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
		Messages: NewMessageDispatcher(),
	}
}

func StringIf(b bool, whenTrue string, whenFalse string) string {
	if b {
		return whenTrue
	}
	return whenFalse
}

type MessageDispatcher struct {
	sync.RWMutex
	subscribers []*subscriber
}

func NewMessageDispatcher() *MessageDispatcher {
	log.Printf("NewMessageDispatcher")
	d := MessageDispatcher{}
	return &d
}

func (d *MessageDispatcher) Send(m appMessage) {
	log.Printf("NewMessageDispatcher.Send to %d", len(d.subscribers))
	d.RLock()
	defer d.RUnlock()
	for _, s := range d.subscribers {
		s.m <- m
	}
}

// Subscribe call onMessage function for each message and return the Unsubscribe function
func (d *MessageDispatcher) Subscribe(onMessage func(m appMessage)) func() {
	log.Printf("NewMessageDispatcher.Subscribe")

	d.Lock()
	defer d.Unlock()
	s := &subscriber{
		m: make(chan appMessage, 1),
	}
	d.subscribers = append(d.subscribers, s)

	go func() {
		for m := range s.m {
			onMessage(m)
		}

	}()

	return func() {
		d.Unsubscribe(s)
	}
}

// Unsubscribe remove the subscriber from the list
func (d *MessageDispatcher) Unsubscribe(s *subscriber) {
	log.Printf("NewMessageDispatcher.Unsubscribe")
	d.Lock()
	defer d.Unlock()
	for i := range d.subscribers {
		if d.subscribers[i] == s {
			close(s.m)
			d.subscribers[i] = d.subscribers[len(d.subscribers)-1]
			d.subscribers[len(d.subscribers)-1] = nil
			d.subscribers = d.subscribers[0 : len(d.subscribers)-1]
			return
		}
	}
}

type subscriber struct {
	// message channel
	m chan appMessage
}

type appMessage struct {
	ID      string
	Class   string
	Stay    bool
	Content app.UI
}
