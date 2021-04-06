package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/simulot/aspiratv/store"
)

func TestProviders(t *testing.T) {
	t.Run("Returns 404 on get with en empty store", func(t *testing.T) {
		s := NewAPIServer(&store.InMemoryStore{})

		request, _ := http.NewRequest(http.MethodGet, "/api/providers/tv", nil)
		response := httptest.NewRecorder()
		s.ServeHTTP(response, request)

		got := response.Result().StatusCode
		want := 404

		if got != want {
			t.Errorf("got status code %v; want %v", got, want)
		}

	})
	t.Run("Has content-type application/json", func(t *testing.T) {
		s := NewAPIServer(&store.InMemoryStore{
			Providers: []store.Provider{
				{
					Name: "tv",
				},
			},
		})

		request, _ := http.NewRequest(http.MethodGet, "/api/providers/tv", nil)
		response := httptest.NewRecorder()
		s.ServeHTTP(response, request)

		got := response.Header().Get("content-type")
		want := "application/json"

		if got != want {
			t.Errorf("got %q; want %q", got, want)
		}

	})

	t.Run("Return a json encoded provider", func(t *testing.T) {
		s := NewAPIServer(&store.InMemoryStore{
			Providers: []store.Provider{
				{
					Name: "tv",
				},
			},
		})

		want := store.Provider{
			Name: "tv",
		}

		request, _ := http.NewRequest(http.MethodGet, "/api/providers/tv", nil)
		response := httptest.NewRecorder()
		s.ServeHTTP(response, request)

		got := store.Provider{}
		err := json.NewDecoder(response.Body).Decode(&got)

		if err != nil {
			t.Fatalf("Got error %q, want nil", err)
			return
		}
		if got != want {
			t.Errorf("got %q; want %q", got, want)
		}

	})

	t.Run("Add a provider missing content-type json/application is rejected", func(t *testing.T) {
		s := NewAPIServer(&store.InMemoryStore{
			Providers: []store.Provider{
				{
					Name: "tv",
				},
			},
		})

		want := store.Provider{
			Name: "tv-new",
		}
		rBody := bytes.NewBuffer(nil)
		err := json.NewEncoder(rBody).Encode(want)
		if err != nil {
			t.Fatalf("Got error when encoding test value %q, want nil", err)
			return
		}
		request, _ := http.NewRequest(http.MethodPost, "/api/providers/", rBody)
		response := httptest.NewRecorder()
		s.ServeHTTP(response, request)

		if response.Result().StatusCode != http.StatusBadRequest {
			t.Fatalf("Got status code %d, want %d", response.Result().StatusCode, http.StatusBadRequest)
			return
		}

	})

	t.Run("Add a provider and return it", func(t *testing.T) {
		s := NewAPIServer(&store.InMemoryStore{
			Providers: []store.Provider{
				{
					Name: "tv",
				},
			},
		})

		want := store.Provider{
			Name: "tv-new",
		}
		rBody := bytes.NewBuffer(nil)
		err := json.NewEncoder(rBody).Encode(want)
		if err != nil {
			t.Fatalf("Got error when encoding test value %q, want nil", err)
			return
		}
		request, _ := http.NewRequest(http.MethodPost, "/api/providers/", rBody)
		request.Header.Set("content-type", "application/json")
		response := httptest.NewRecorder()
		s.ServeHTTP(response, request)

		if response.Result().StatusCode != http.StatusAccepted {
			t.Fatalf("Got status code %d, want %d", response.Result().StatusCode, http.StatusAccepted)
			return
		}

		got := store.Provider{}
		err = json.NewDecoder(response.Body).Decode(&got)

		if err != nil {
			t.Fatalf("Got error %q, want nil", err)
			return
		}
		if got != want {
			t.Errorf("got %q; want %q", got, want)
		}

	})

	t.Run("Get a list of providers", func(t *testing.T) {
		want := []store.Provider{
			{
				Name: "tv1",
			},
			{
				Name: "tv2",
			},
			{
				Name: "tv3",
			},
		}
		s := NewAPIServer(&store.InMemoryStore{
			Providers: want,
		})

		request, _ := http.NewRequest(http.MethodGet, "/api/providers/", nil)
		response := httptest.NewRecorder()
		s.ServeHTTP(response, request)

		if response.Result().StatusCode != http.StatusOK {
			t.Fatalf("Got status code %d, want %d", response.Result().StatusCode, http.StatusOK)
			return
		}

		got := []store.Provider{}
		err := json.NewDecoder(response.Body).Decode(&got)

		if err != nil {
			t.Fatalf("Got error %q, want nil", err)
			return
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %q; want %q", got, want)
		}

	})
}

type cancelableProvider struct {
	store.InMemoryStore
	cancelled bool
}

func (s *cancelableProvider) GetProvider(ctx context.Context, name string) (store.Provider, error) {
	type result struct {
		p   store.Provider
		err error
	}

	data := make(chan result, 1)
	go func() {
		var r result

		time.Sleep(10 * time.Millisecond)
		r.p, r.err = s.InMemoryStore.GetProvider(ctx, name)
		data <- r
	}()

	select {
	case r := <-data:
		return r.p, r.err

	case <-ctx.Done():
		s.cancelled = true
		return store.Provider{}, ctx.Err()
	}
}

func TestCancel(t *testing.T) {
	t.Run("Tell store to cancel work when request is cancelled", func(t *testing.T) {
		st := cancelableProvider{
			store.InMemoryStore{
				Providers: []store.Provider{
					{
						Name: "tv",
					},
				},
			},
			false,
		}
		s := NewAPIServer(&st)

		request, _ := http.NewRequest(http.MethodGet, "/api/providers/tv", nil)

		cancellingCtx, cancel := context.WithCancel(request.Context())
		defer cancel()
		time.AfterFunc(5*time.Millisecond, cancel)

		request = request.WithContext(cancellingCtx)

		response := httptest.NewRecorder()
		s.ServeHTTP(response, request)

		if !st.cancelled {
			t.Errorf("Store was not told to cancel")
		}
	})
	t.Run("Server must return 503 Service Unavailable when the client cancels the requestest", func(t *testing.T) {
		st := cancelableProvider{
			store.InMemoryStore{
				Providers: []store.Provider{
					{
						Name: "tv",
					},
				},
			},
			false,
		}
		s := NewAPIServer(&st)

		request, _ := http.NewRequest(http.MethodGet, "/api/providers/tv", nil)

		cancellingCtx, cancel := context.WithCancel(request.Context())
		defer cancel()
		time.AfterFunc(5*time.Millisecond, cancel)

		request = request.WithContext(cancellingCtx)

		response := httptest.NewRecorder()
		s.ServeHTTP(response, request)

		if response.Code != http.StatusServiceUnavailable {
			t.Errorf("get status %v, want %v", response.Code, http.StatusServiceUnavailable)
		}
	})
}
