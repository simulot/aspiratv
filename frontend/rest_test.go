package frontend

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/simulot/aspiratv/backend"
	"github.com/simulot/aspiratv/store"
)

func TestRestStore(t *testing.T) {

	t.Run("Call the rest api with empty store, no error expected", func(t *testing.T) {
		spy := newSpyStore(t, nil)
		s, tearDownSrv := setupApiServer(t, spy)
		defer tearDownSrv()
		ctx := context.Background()

		restStore := NewRestStore(s.URL + "/api/")
		_, err := restStore.GetProviderList(ctx)

		if spy.GetProviderListCalled == 0 {
			t.Error("GetProviderList must be called")
		}
		if err != nil {
			t.Errorf("Unexpected error %s", err)
		}
	})
	t.Run("Call rest.GetProvider, should find it", func(t *testing.T) {
		spy := newSpyStore(t, []store.Provider{
			{
				Name: "tv-1",
			},
			{
				Name: "tv-2",
			},
		})
		s, tearDownSrv := setupApiServer(t, spy)
		defer tearDownSrv()
		ctx := context.Background()

		restStore := NewRestStore(s.URL + "/api/")
		got, err := restStore.GetProvider(ctx, "tv-2")

		if spy.GetProviderCalled == 0 {
			t.Error("GetProvider must be called")
		}
		if err != nil {
			t.Errorf("Unexpected error %s", err)
		}

		want := spy.InMemoryStore.Providers[1]

		if got != want {
			t.Errorf("Want %+v, got %+v", want, got)
		}
	})

	t.Run("Call rest.GetProvider, should not find it", func(t *testing.T) {
		spy := newSpyStore(t, []store.Provider{
			{
				Name: "tv-1",
			},
			{
				Name: "tv-2",
			},
		})
		s, tearDownSrv := setupApiServer(t, spy)
		defer tearDownSrv()
		ctx := context.Background()

		restStore := NewRestStore(s.URL + "/api/")
		got, err := restStore.GetProvider(ctx, "tv-3")

		if spy.GetProviderCalled == 0 {
			t.Error("GetProvider must be called")
		}
		if err == nil {
			t.Errorf("Expecting an error, but got no error")
		}
		var hErr httpError
		if !errors.As(err, &hErr) {
			t.Errorf("Expecting href error, but got something else: %s", err)
		} else {
			got := hErr.StatusCode
			want := http.StatusNotFound
			if got != want {
				t.Errorf("Exprecting status code %d, got %d", want, got)
			}
		}
		want := store.Provider{}

		if got != want {
			t.Errorf("Want %+v, got %+v", want, got)
		}
	})

	t.Run("Call rest.SetProvider, should store it", func(t *testing.T) {
		spy := newSpyStore(t, []store.Provider{
			{
				Name: "tv-1",
			},
			{
				Name: "tv-2",
			},
		})
		s, tearDownSrv := setupApiServer(t, spy)
		defer tearDownSrv()
		ctx := context.Background()

		newP := store.Provider{
			Name: "tv-3",
		}

		restStore := NewRestStore(s.URL + "/api/")
		got, err := restStore.SetProvider(ctx, newP)

		if spy.SetProviderCalled == 0 {
			t.Error("SetProvider must be called")
		}
		if err != nil {
			t.Errorf("Expecting no error, but got error: %s", err)
		}
		want := store.Provider{Name: "tv-3"}

		if got != want {
			t.Errorf("Want %+v, got %+v", want, got)
		}

		found := false
		for i := range spy.Providers {
			if spy.Providers[i].Name == "tv-3" {
				found = true
			}
		}

		if !found {
			t.Errorf("Newly set provider is not found in store")
		}

	})

	t.Run("Call rest.Search and get results", func(t *testing.T) {
		spy := newSpyStore(t, []store.Provider{
			{
				Name: "tv-1",
			},
		})
		spy.searchResults = make([]store.SearchResult, 5)
		s, tearDownSrv := setupApiServer(t, spy)
		defer tearDownSrv()
		ctx := context.Background()
		restStore := NewRestStore(wsURL(t, s.URL) + "/api/")

		results, err := restStore.Search(ctx, store.SearchQuery{Title: "Hello"})
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
	})
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

type spyStore struct {
	store.InMemoryStore

	searchResults         []store.SearchResult
	delay                 time.Duration
	recordSent            int
	cancelled             bool
	GetProviderListCalled int
	GetProviderCalled     int
	SetProviderCalled     int
	SearchCalled          int
	start                 time.Time
	searchQuery           store.SearchQuery

	t *testing.T
}

func newSpyStore(t *testing.T, providers []store.Provider) *spyStore {
	return &spyStore{
		InMemoryStore: store.InMemoryStore{
			Providers: providers,
		},
		t:     t,
		start: time.Now(),
	}
}

func (s *spyStore) GetProviderList(ctx context.Context) ([]store.Provider, error) {
	s.GetProviderListCalled++
	return s.InMemoryStore.GetProviderList(ctx)
}

func (s *spyStore) GetProvider(ctx context.Context, name string) (store.Provider, error) {
	s.GetProviderCalled++
	return s.InMemoryStore.GetProvider(ctx, name)
}

func (s *spyStore) SetProvider(ctx context.Context, p store.Provider) (store.Provider, error) {
	s.SetProviderCalled++
	return s.InMemoryStore.SetProvider(ctx, p)
}

func (s *spyStore) Search(ctx context.Context, q store.SearchQuery) (<-chan store.SearchResult, error) {
	s.SearchCalled++
	s.searchQuery = q
	c := make(chan store.SearchResult, 1)
	go func() {
		defer close(c)
		for _, r := range s.searchResults {
			select {
			case <-ctx.Done():
				s.cancelled = true
				return
			default:
				time.Sleep(s.delay)
				c <- r
				s.recordSent++
			}
		}
	}()

	return c, nil
}

func setupApiServer(t *testing.T, st store.Store) (*httptest.Server, func()) {
	t.Helper()
	s := httptest.NewServer(backend.NewAPIServer(st))
	return s, func() {}
}
