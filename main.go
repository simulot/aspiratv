package main

import (
	"context"
	"log"
	"net/http"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/simulot/aspiratv/pkg/backend"
	"github.com/simulot/aspiratv/pkg/frontend"
	"github.com/simulot/aspiratv/pkg/myhttp"
	"github.com/simulot/aspiratv/pkg/providers"
	"github.com/simulot/aspiratv/pkg/providers/artetv"
	"github.com/simulot/aspiratv/pkg/store"
)

func main() {

	ctx := context.Background()
	// TODO: Handle gracefull shutdown

	frontend.InitializeWebApp(context.Background())
	app.Route("/", &frontend.LandingPage{})
	app.Route("/search", &frontend.Search{})
	app.Route("/settings", &frontend.Settings{})
	app.Route("/credits", &frontend.Credits{})
	app.RunWhenOnBrowser()

	// Starting here, the server side

	serverAddress := "localhost:8000"
	st := store.NewStoreJSON("config.json")

	providers := []providers.Provider{
		// mockup.NewMockup(),
		artetv.NewArte(
			artetv.WithClientConfigurations(
				myhttp.WithRequestLogger(log.Default()),
				myhttp.WithResponseLogger(
					myhttp.NewPayloadDumper(log.Default(), "tmp", "arte_*.json", myhttp.JSONDumper),
				),
			),
		),
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
