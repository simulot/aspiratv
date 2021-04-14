package frontend

import (
	"context"
	"fmt"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/simulot/aspiratv/backend"
	"github.com/simulot/aspiratv/models"
	"github.com/simulot/aspiratv/providers"
	"github.com/simulot/aspiratv/store"
)

func TestRestSearch(t *testing.T) {
	t.Run("Call rest.Search and get results", func(t *testing.T) {
		spy := spyProvider{
			searchResults: make([]models.SearchResult, 100),
		}
		s, tearDownSrv := setupApiServer(t, &spy)
		defer tearDownSrv()
		ctx := context.Background()
		restStore := NewRestStore(wsURL(t, s.URL) + "/api/")

		q := models.SearchQuery{Title: "Hello"}
		results, err := restStore.Search(ctx, q)
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
			return
		}

		got := 0
		for range results {
			got++
		}
		want := len(spy.searchResults)
		if got != want {
			t.Errorf("Expecting %d result, got %d", want, got)
		}
		if q != spy.searchQuery {
			t.Errorf("Got %v when expecting %v", spy.searchQuery, q)
		}
	})

	t.Run("Call rest.Search and cancel it", func(t *testing.T) {
		spy := spyProvider{
			searchDelay:   10 * time.Millisecond,
			searchResults: make([]models.SearchResult, 100),
		}
		s, tearDownSrv := setupApiServer(t, &spy)
		defer tearDownSrv()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		restStore := NewRestStore(wsURL(t, s.URL) + "/api/")

		results, err := restStore.Search(ctx, models.SearchQuery{Title: "Hello"})
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
			return
		}

		imDone := make(chan struct{})
		time.AfterFunc(50*time.Millisecond, func() {
			cancel()
		})

		got := 0
		go func() {
			defer close(imDone)
			for {
				select {
				case <-ctx.Done():
					return
				case _, ok := <-results:
					if !ok {
						return
					}
					got++
				}
			}
		}()

		<-imDone

		want := len(spy.searchResults)
		if got == 0 {
			t.Errorf("Expecting more than %d result", got)
		}
		if got == want {
			t.Errorf("Expecting less then %d result, got %d", want, got)
		}
	})

}

func setupApiServer(t *testing.T, spy providers.Provider) (*httptest.Server, func()) {
	t.Helper()
	s := httptest.NewServer(backend.NewServer(&store.InMemoryStore{}, []providers.Provider{spy}))
	return s, func() {}
}

type spyProvider struct {
	describeCalled  bool
	describe        providers.ProviderDescription
	searchCalled    bool
	searchResults   []models.SearchResult
	searchQuery     models.SearchQuery
	searchCancelled bool
	searchDelay     time.Duration
	searchSent      int
}

func (s *spyProvider) ProviderDescribe(ctx context.Context) providers.ProviderDescription {
	s.describeCalled = true
	return s.describe
}

func (s *spyProvider) Search(ctx context.Context, q models.SearchQuery) (<-chan models.SearchResult, error) {
	s.searchCalled = true
	s.searchQuery = q
	results := make(chan models.SearchResult, 1)
	go func() {
		defer func() {
			close(results)
		}()
		for _, r := range s.searchResults {
			select {
			case <-ctx.Done():
				s.searchCancelled = true
				return
			case results <- r:
				s.searchSent++
				time.Sleep(s.searchDelay)
			}
		}
	}()

	return results, nil
}
func (s *spyProvider) makeFakeResults(howMany int) {
	num := 0
	for ; howMany > 0; howMany-- {
		s.searchResults = append(s.searchResults, models.SearchResult{
			Title: fmt.Sprintf("Result #%d", num)},
		)
		num++
	}
}

func wsURL(t *testing.T, s string) string {
	t.Helper()
	u, err := url.Parse(s)
	if err != nil {
		t.Errorf("Can't parse url: %s", err)
		return ""
	}
	u.Scheme = "ws"
	return u.String()
}
