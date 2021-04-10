package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/simulot/aspiratv/frontend"

	"github.com/maxence-charriere/go-app/v8/pkg/app"
	"github.com/simulot/aspiratv/backend"

	"github.com/simulot/aspiratv/store"
)

func main() {

	// Initialize web application storage and state
	log.Printf("%+#v", os.Args)
	frontend.InitializeWebApp()
	app.Route("/", &frontend.LandingPage{})
	app.Route("/search", &frontend.SearchOnline{})
	app.Route("/credits", &frontend.Credits{})
	app.RunWhenOnBrowser()

	// Starting here, the server side

	st := &dummyStore{
		InMemoryStore: store.InMemoryStore{
			Providers: []store.Provider{
				{Name: "TV1"},
				{Name: "TV2"},
			},
		},
	}

	mux := http.NewServeMux()
	mux.Handle("/", &app.Handler{
		Name:        "Hello",
		Description: "An Hello World! example",
		Styles: []string{
			"web/mystyles.css",
		},
	})

	mux.Handle(backend.APIURL, backend.NewAPIServer(st))
	if err := http.ListenAndServe(":8000", mux); err != nil {
		log.Fatal(err)
	}
}

type dummyStore struct {
	store.InMemoryStore
}

func (s *dummyStore) Search(ctx context.Context) (<-chan store.SearchResult, error) {
	c := make(chan store.SearchResult, 1)
	go func() {
		defer close(c)
		start := time.Now()
		rand.Seed(time.Now().Unix())
		itemsNubmer := rand.Intn(20)
		log.Printf("Search gets %d records", itemsNubmer)
		for i := 1; i <= itemsNubmer; i++ {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(time.Duration(rand.Intn(10)) * 100 * time.Millisecond)
				c <- store.SearchResult{
					Num:   i,
					Title: fmt.Sprintf("Item %d at %s", i, time.Since(start)),
				}
			}
		}
	}()

	return c, nil
}
