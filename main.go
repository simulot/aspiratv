package main

import (
	"context"
	"log"
	"net/http"
	"net/url"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/simulot/aspiratv/pkg/backend"
	"github.com/simulot/aspiratv/pkg/frontend"
	"github.com/simulot/aspiratv/pkg/providers"
	"github.com/simulot/aspiratv/pkg/providers/mockup"
	"github.com/simulot/aspiratv/pkg/store"
)

func main() {

	// TODO: Handle gracefull shutdown
	ctx := context.Background()

	var st store.Store
	var endpoint string
	var u *url.URL
	if app.IsClient {
		u = app.Window().URL()
		u.Scheme = "http"
		u.Path = "/api/"
		u.RawQuery = ""
		log.Printf("[CLIENT] API endpoint: %s", u.String())
		endpoint = u.String()
		st = store.NewRestStore(endpoint)
	} else {
		st = store.NewJSONStore("config.json")
		u = app.Window().URL()
		u.Scheme = "http"
		u.Host = "localhost:8000"
		u.Path = "/api/"
		log.Printf("[SERVER] API endpoint: %s", u.String())
		endpoint = u.String()
	}
	frontend.InitializeWebApp(ctx, endpoint, st)

	app.Route("/", &frontend.MyApp{})
	// app.Route("/search", &frontend.SearchPage{})
	// app.Route("/subscriptions", &frontend.SubscriptionPage{})
	// app.Route("/settings", &frontend.SettingsPage{})
	// app.Route("/credits", &frontend.CreditsPage{})
	app.RunWhenOnBrowser()

	// Starting here, the server side

	serverAddress := "localhost:8000"

	providers := []providers.Provider{
		mockup.NewMockup(),
		// artetv.NewArte(
		// 	artetv.WithClientConfigurations(
		// 		myhttp.WithRequestLogger(log.Default()),
		// 		myhttp.WithResponseLogger(
		// 			myhttp.NewPayloadDumper(log.Default(), "testdata", "arte_*.json", myhttp.JSONDumper),
		// 		),
		// 	),
		// ),
		// francetv.NewFranceTV(
		// 	francetv.WithClientConfigurations(
		// 		myhttp.WithRequestLogger(log.Default()),
		// 		myhttp.WithResponseLogger(
		// 			myhttp.NewPayloadDumper(log.Default(), "tmp", "francetv_*.json", func(b []byte) []byte {
		// 				var s string
		// 				err := json.Unmarshal(b, &s)
		// 				if err != nil {
		// 					return b
		// 				}+
		// 				return []byte(s)
		// 			})),
		// 	),
		// ),
	}

	mux := http.NewServeMux()
	mux.Handle("/", logRequests(&app.Handler{
		Name:        "Hello",
		Description: "An Hello World! example",
		Styles: []string{
			"web/mystyles.css",
			"https://cdn.jsdelivr.net/npm/@mdi/font@5.9.55/css/materialdesignicons.min.css",
		},
	}))
	mux.Handle(backend.APIURL, logRequests(backend.NewServer(ctx, st, providers)))

	if err := http.ListenAndServe(serverAddress, mux); err != nil {
		log.Fatal(err)
	}
}

func logRequests(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[HTTPSERVER] %s %s", r.Method, r.URL.String())
		h.ServeHTTP(w, r)
	}
}
