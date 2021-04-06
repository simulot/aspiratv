package backend

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/simulot/aspiratv/store"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type searchSpyStore struct {
	store.InMemoryStore

	results    []store.SearchResult
	delay      time.Duration
	recordSent int
	cancelled  bool
	called     bool
	start      time.Time

	t *testing.T
}

func newSearchSpyStore(t *testing.T, providers []store.Provider) *searchSpyStore {
	return &searchSpyStore{
		InMemoryStore: store.InMemoryStore{
			Providers: providers,
		},
		t:     t,
		start: time.Now(),
	}
}

// func (s *searchSpyStore) at() time.Duration {
// 	return time.Since(s.start)
// }

func (s *searchSpyStore) makeFakeResults(howMany int) {
	ballast := strings.Repeat("*", 512)
	num := 0
	for ; howMany > 0; howMany-- {
		s.results = append(s.results, store.SearchResult{
			Num:     num,
			Ballast: ballast})
		num++
	}
}

func (s *searchSpyStore) Search(ctx context.Context) (<-chan store.SearchResult, error) {
	s.called = true
	c := make(chan store.SearchResult, 1)
	go func() {
		defer close(c)
		for _, r := range s.results {
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

func TestSearch(t *testing.T) {

	t.Run("/api/search should open a websocket and call Search", func(t *testing.T) {
		st := newSearchSpyStore(t, nil)
		s := httptest.NewServer(NewAPIServer(st))
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		c, _, err := websocket.Dial(ctx, s.URL+"/api/search/", nil)

		if !st.called {
			t.Logf("Search function not called")
		}
		var closeErr websocket.CloseError
		if err != nil && errors.As(err, &closeErr) {
			if closeErr.Reason != "no more result" {
				t.Errorf("Got Close error with message %q, want %q", closeErr.Reason, "no more result")
			} else {
				t.Fatalf("Got error from Dial that isn't CloseError: %q", err)
			}
		}
		defer c.Close(websocket.StatusInternalError, "the sky is falling")

	})

	t.Run("/api/search should return the expected records and check no more result", func(t *testing.T) {
		st := newSearchSpyStore(t, nil)
		st.makeFakeResults(100)
		s := httptest.NewServer(NewAPIServer(st))
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		c, _, err := websocket.Dial(ctx, s.URL+"/api/search/", nil)
		if err != nil {
			t.Fatalf("websocket.Dial error: %v", err)
			return
		}
		defer c.Close(websocket.StatusInternalError, "the sky is falling")

		wsjson.Write(ctx, c, "my search query")

		iamDone := make(chan struct{})
		got := 0
		go func(t *testing.T) {
			defer close(iamDone)
			for {
				select {
				case <-ctx.Done():
					t.Error("Done received, not expected")
					return
				default:
					var closeErr websocket.CloseError

					var result store.SearchResult
					err := wsjson.Read(ctx, c, &result)
					if err != nil {

						if err != nil && errors.As(err, &closeErr) {
							if closeErr.Code != websocket.StatusNormalClosure {
								t.Errorf("Got Close.Code %v, want %v", closeErr.Code, websocket.StatusNormalClosure)
								return
							}
							if closeErr.Reason != "no more result" {
								t.Errorf("Got Close error with message %q, want %q", closeErr.Reason, "no more result")
								return
							}
						}
						if err != nil && closeErr.Reason == "no more result" {
							c.Close(websocket.StatusNormalClosure, "")
							return
						}
					}
					got++
				}
			}
		}(t)
		<-iamDone
		if got != len(st.results) {
			t.Errorf("Got %v results, expected %v", got, len(st.results))
		}
		if st.cancelled {
			t.Errorf("Search been cancelled unexpectedly")

		}
	})

	t.Run("/api/search should be cancelled when the client cancels", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		st := newSearchSpyStore(t, nil)
		st.makeFakeResults(100)
		st.delay = 10 * time.Millisecond

		s := httptest.NewServer(NewAPIServer(st))
		defer cancel()

		c, _, err := websocket.Dial(ctx, s.URL+"/api/search/", nil)
		if err != nil {
			t.Fatalf("websocket.Dial error: %v", err)
			return
		}
		defer c.Close(websocket.StatusInternalError, "the sky is falling")

		if err != nil {
			t.Fatalf("websocket.Dial error: %v", err)
			return
		}

		time.AfterFunc(50*time.Millisecond, func() {
			cancel()
		})

		iamDone := make(chan struct{})
		got := 0
		go func(t *testing.T) {
			defer close(iamDone)
			for {
				select {
				case <-ctx.Done():
					return
				default:
					var result store.SearchResult
					err := wsjson.Read(ctx, c, &result)
					if err != nil {
						return
					}
					got++
				}
			}
		}(t)
		<-iamDone
		c.Close(websocket.StatusNormalClosure, "All good")
		ctx = c.CloseRead(context.Background())
		<-ctx.Done()
		if !(got < len(st.results)) {
			t.Errorf("Search not been cancelled")

		}
	})
}
