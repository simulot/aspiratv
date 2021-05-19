package frontend

import (
	"context"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/simulot/aspiratv/pkg/backend"
	"github.com/simulot/aspiratv/pkg/models"
	"github.com/simulot/aspiratv/pkg/providers"
	"github.com/simulot/aspiratv/pkg/store"
)

func TestRestSearch(t *testing.T) {
	t.Run("Call rest.Search and get results", func(t *testing.T) {
		spy := spyProvider{
			searchResults: make([]models.SearchResult, 100),
		}
		s, tearDownSrv := setupApiServer(t, &spyStore{}, &spy)
		defer tearDownSrv()
		ctx := context.Background()
		restStore := NewRestStore(wsURL(t, s.URL)+"/api/", nil)

		q := models.SearchQuery{Title: "Hello", AiredAfter: time.Now()}
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
		if q.Title != spy.searchQuery.Title {
			t.Errorf("Got %v when expecting %v", spy.searchQuery, q)
		}
	})

	t.Run("Call rest.Search and cancel it", func(t *testing.T) {
		spy := spyProvider{
			searchDelay:   10 * time.Millisecond,
			searchResults: make([]models.SearchResult, 100),
		}
		s, tearDownSrv := setupApiServer(t, &spyStore{}, &spy)
		defer tearDownSrv()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		restStore := NewRestStore(wsURL(t, s.URL)+"/api/", nil)

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

func TestSettings(t *testing.T) {
	t.Run("Test GetSettings", func(t *testing.T) {
		spySt := spyStore{
			settings: models.Settings{
				LibraryPath: "mypath",
			},
		}
		s, tearDownSrv := setupApiServer(t, &spySt, nil)

		defer tearDownSrv()
		ctx := context.Background()
		restStore := NewRestStore(s.URL+"/api/", nil)

		got, err := restStore.GetSettings(ctx)
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
			return
		}

		want := spySt.settings
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Expecting %#v result, got %#v", want, got)
		}
	})

}

func setupApiServer(t *testing.T, st store.Store, spy providers.Provider) (*httptest.Server, func()) {
	t.Helper()
	s := httptest.NewServer(backend.NewServer(context.TODO(), st, []providers.Provider{spy}))
	return s, func() {}
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

func (s *spyProvider) GetMedias(ctx context.Context, task models.DownloadTask) (<-chan models.DownloadItem, error) {
	return nil, nil
}
func (s *spyProvider) Name() string { return "spy" }
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

// func (s *spyProvider) makeFakeResults(howMany int) {
// 	num := 0
// 	for ; howMany > 0; howMany-- {
// 		s.searchResults = append(s.searchResults, models.SearchResult{
// 			Title: fmt.Sprintf("Result #%d", num)},
// 		)
// 		num++
// 	}
// }

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
	settings          models.Settings
	getSettingsCalled bool
	setSettingsCalled bool
}

func (s *spyStore) GetSubscription(ctx context.Context, UUID uuid.UUID) (models.Subscription, error) {
	return models.Subscription{}, nil
}
func (s *spyStore) GetAllSubscriptions(ctx context.Context) ([]models.Subscription, error) {
	return []models.Subscription{}, nil
}
func (s *spyStore) SetSubscription(context.Context, models.Subscription) (models.Subscription, error) {
	return models.Subscription{}, nil
}

func (s *spyStore) GetSettings(ctx context.Context) (models.Settings, error) {
	s.getSettingsCalled = true
	return s.settings, nil
}

func (s *spyStore) SetSettings(ctx context.Context, settings models.Settings) (models.Settings, error) {
	s.setSettingsCalled = true
	return settings, nil
}
