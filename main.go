package main

import (
	"log"
	"net/http"
	"os"

	"github.com/simulot/aspiratv/frontend"
	"github.com/simulot/aspiratv/providers"
	"github.com/simulot/aspiratv/providers/artetv"

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

	providers := []providers.Provider{
		artetv.NewArte(),
	}
	st := store.InMemoryStore{}

	mux := http.NewServeMux()
	mux.Handle("/", &app.Handler{
		Name:        "Hello",
		Description: "An Hello World! example",
		Styles: []string{
			"web/mystyles.css",
		},
	})

	mux.Handle(backend.APIURL, backend.NewServer(&st, providers))
	if err := http.ListenAndServe(":8000", mux); err != nil {
		log.Fatal(err)
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
