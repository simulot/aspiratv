package frontend

import (
	"github.com/maxence-charriere/go-app/v8/pkg/app"
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
	s           *RestClient //access to the backend store using RESP API
	currentPage PageID
	menuItems   []Menuitem
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
	}
}

func StringIf(b bool, whenTrue string, whenFalse string) string {
	if b {
		return whenTrue
	}
	return whenFalse
}
