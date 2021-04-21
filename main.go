package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/maxence-charriere/go-app/v8/pkg/app"
	"github.com/simulot/aspiratv/backend"
	"github.com/simulot/aspiratv/frontend"
	"github.com/simulot/aspiratv/myhttp"
	"github.com/simulot/aspiratv/providers"
	"github.com/simulot/aspiratv/providers/artetv"
	"github.com/simulot/aspiratv/providers/francetv"
	"github.com/simulot/aspiratv/store"
)

func main() {
	// Initialize web application storage and state
	frontend.InitializeWebApp()
	app.Route("/", &frontend.LandingPage{})
	app.Route("/search", &frontend.SearchOnline{})
	app.Route("/credits", &frontend.Credits{})
	app.RunWhenOnBrowser()

	// Starting here, the server side

	providers := []providers.Provider{
		artetv.NewArte(
			artetv.WithClientConfigurations(
				myhttp.WithRequestLogger(log.Default()),
				myhttp.WithResponseLogger(
					myhttp.NewPayloadDumper(log.Default(), "tmp", "arte_*.json", func(b []byte) []byte { return b }),
				),
			),
		),
		francetv.NewFranceTV(
			francetv.WithClientConfigurations(
				myhttp.WithRequestLogger(log.Default()),
				myhttp.WithResponseLogger(
					myhttp.NewPayloadDumper(log.Default(), "tmp", "francetv_*.json", func(b []byte) []byte {
						var s string
						err := json.Unmarshal(b, &s)
						if err != nil {
							return b
						}
						return []byte(s)
					})),
			),
		),
	}
	st := store.InMemoryStore{}

	mux := http.NewServeMux()
	mux.Handle("/", logRequests(&app.Handler{
		Name:        "Hello",
		Description: "An Hello World! example",
		Styles: []string{
			"web/mystyles.css",
			"https://cdn.jsdelivr.net/npm/@mdi/font@5.9.55/css/materialdesignicons.min.css",
		},
	}))
	mux.Handle(backend.APIURL, logRequests(backend.NewServer(&st, providers)))

	if err := http.ListenAndServe(":8000", mux); err != nil {
		log.Fatal(err)
	}
}

func yo(s string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, s)
	}
}

func logRequests(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[HTTPSERVER] %s %s", r.Method, r.URL.String())
		h.ServeHTTP(w, r)
	}
}

/*
type dummyStore struct {
	store.InMemoryStore
}

func (s *dummyStore) Search(ctx context.Context, q models.SearchQuery) (<-chan models.SearchResult, error) {
	c := make(chan models.SearchResult, 1)
	go func() {
		defer close(c)
		start := time.Now()
		rand.Seed(time.Now().Unix())
		itemsNubmer := rand.Intn(19) + 1
		log.Printf("Search gets %d records", itemsNubmer)
		for i := 1; i <= itemsNubmer; i++ {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(time.Duration(rand.Intn(10)) * 100 * time.Millisecond)
				c <- models.SearchResult{
					Title: fmt.Sprintf("Item %s(%d) at %s", q.Title, i, time.Since(start)),
				}
			}
		}
	}()

	return c, nil
}
*/
