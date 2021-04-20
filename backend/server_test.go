package backend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/simulot/aspiratv/models"
	"github.com/simulot/aspiratv/providers"
	"github.com/simulot/aspiratv/store"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

/*
	Test of API Handlers
	Functional tests are done when testing client
*/

func TestProviderDescribleHandler(t *testing.T) {
	t.Run("DescribeProvider called", func(t *testing.T) {
		spy := &spyProvider{}
		s := NewServer(&store.InMemoryStore{}, []providers.Provider{spy})

		request, _ := http.NewRequest(http.MethodGet, "/api/providers/", nil)
		response := httptest.NewRecorder()
		s.ServeHTTP(response, request)

		statusGot := response.Result().StatusCode
		statusWant := 200

		if statusGot != statusWant {
			t.Errorf("got status code %v; want %v", statusGot, statusWant)
		}

		if !spy.describeCalled {
			t.Error("store's ProvidersDescribe not called")
		}

		typeGot := response.Header().Get("content-type")
		typeWant := "application/json"
		if typeGot != typeWant {
			t.Errorf("got %q; want %q", typeGot, typeWant)
		}

		got := []providers.Description{}
		b, err := io.ReadAll(response.Body)
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
			return
		}
		err = json.Unmarshal(b, &got)
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
			return
		}
		if !reflect.DeepEqual(got[0], spy.describe) {
			t.Errorf("Got %+#v, want %+#v", got, spy.describe)
		}
	})
}

func TestSearchHandler(t *testing.T) {

	t.Run("/api/search should return the expected records and check for no more result status", func(t *testing.T) {
		spy := &spyProvider{}
		srv := NewServer(&store.InMemoryStore{}, []providers.Provider{spy})

		spy.makeFakeResults(100)
		s := httptest.NewServer(srv)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		c, _, err := websocket.Dial(ctx, wsURL(t, s.URL)+"/api/search/", nil)
		if err != nil {
			t.Fatalf("websocket.Dial error: %v", err)
			return
		}
		defer c.Close(websocket.StatusInternalError, "the sky is falling")

		err = wsjson.Write(ctx, c, models.SearchQuery{Title: "my search query"})
		if err != nil {
			t.Errorf("websocket.Write error: %v", err)
			return
		}

		var status string
		err = wsjson.Read(ctx, c, &status)

		if err != nil {
			t.Logf("Unexpected error:%s", err)
		}
		if status != "OK" {
			t.Fatalf("Invalid status: %s", status)
		}

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

					var result models.SearchResult
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

		if spy.searchQuery.Title != "my search query" {
			t.Errorf("Expecting Query.Title %q, got %q", "my search query", spy.searchQuery.Title)
		}

		if got != len(spy.searchResults) {
			t.Errorf("Got %v results, expected %v", got, len(spy.searchResults))
		}
		if spy.searchCancelled {
			t.Errorf("Search been cancelled unexpectedly")

		}
	})

	t.Run("/api/search should handle request cancellation", func(t *testing.T) {
		spy := &spyProvider{}
		srv := NewServer(&store.InMemoryStore{}, []providers.Provider{spy})

		spy.makeFakeResults(100)
		spy.searchDelay = 10 * time.Millisecond
		s := httptest.NewServer(srv)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		c, _, err := websocket.Dial(ctx, wsURL(t, s.URL)+"/api/search/", nil)
		if err != nil {
			t.Fatalf("websocket.Dial error: %v", err)
			return
		}
		defer c.Close(websocket.StatusInternalError, "the sky is falling")

		err = wsjson.Write(ctx, c, models.SearchQuery{Title: "my search query"})
		if err != nil {
			t.Errorf("websocket.Write error: %v", err)
			return
		}

		var status string
		err = wsjson.Read(ctx, c, &status)

		if err != nil {
			t.Logf("Unexpected error:%s", err)
		}
		if status != "OK" {
			t.Fatalf("Invalid status: %s", status)
		}

		time.AfterFunc(20*time.Millisecond, func() {
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
					var closeErr websocket.CloseError

					var result models.SearchResult
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

		if got == 0 {
			t.Errorf("Got 0 results when expecting more")
		}
		if got == len(spy.searchResults) {
			t.Errorf("Got %v results, expected much less", got)
		}

		/* Can't see cancellation
		if !st.searchCancelled {
			t.Errorf("Search not been cancelled")
		}
		*/

		if spy.searchSent != got {
			t.Logf("Got %d results, but %d was sent", got, spy.searchSent)
		}

		if spy.searchSent == len(spy.searchResults) {
			t.Errorf("Search sent %d result despite cancellation request", spy.searchSent)

		}
	})
}

type spyProvider struct {
	describeCalled  bool
	describe        providers.Description
	searchCalled    bool
	searchResults   []models.SearchResult
	searchQuery     models.SearchQuery
	searchCancelled bool
	searchDelay     time.Duration
	searchSent      int
}

func (s *spyProvider) ProviderDescribe(ctx context.Context) providers.Description {
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
