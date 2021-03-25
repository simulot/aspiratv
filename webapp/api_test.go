package webapp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

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

		if response.Result().StatusCode != http.StatusFound {
			t.Fatalf("Got status code %d, want %d", response.Result().StatusCode, http.StatusFound)
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
