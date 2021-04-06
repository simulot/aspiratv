package main

import (
	"log"
	"net/http"

	"github.com/simulot/aspiratv/frontend"

	"github.com/maxence-charriere/go-app/v8/pkg/app"
	"github.com/simulot/aspiratv/backend"

	"github.com/simulot/aspiratv/store"
)

func main() {

	app.Route("/", &frontend.MyApp{})
	app.RunWhenOnBrowser()

	st := &store.InMemoryStore{Providers: []store.Provider{
		{Name: "TV1"},
		{Name: "TV2"},
	}}
	mux := http.NewServeMux()

	// Finally, launching the server that serves the app is done by using the Go
	// standard HTTP package.
	//
	// The Handler is an HTTP handler that serves the client and all its
	// required resources to make it work into a web browser. Here it is
	// configured to handle requests with a path that starts with "/".
	mux.Handle("/", &app.Handler{
		Name:        "Hello",
		Description: "An Hello World! example",
	})

	mux.Handle(backend.APIURL, backend.NewAPIServer(st))
	if err := http.ListenAndServe(":8000", mux); err != nil {
		log.Fatal(err)
	}
}
