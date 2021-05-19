package library

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/simulot/aspiratv/pkg/download"
	"github.com/simulot/aspiratv/pkg/models"
)

func TestDowload(t *testing.T) {
	os.RemoveAll("data")
	t.Run("Happy path", func(t *testing.T) {
		task := models.DownloadTask{Result: models.SearchResult{Title: "Test", Type: models.TypeSeries}}
		dl := newFakeDownloader(func(s *fakeDownloader) error { return nil })
		c := makeMediaChannel(t, task, 1, dl)
		NewBatchDownloader(task.Result.Show, "data", models.DefaultFileNamer[models.PathNamingType(task.Result.Type)]).
			Download(context.
				Background(),
				c)

		if !dl.called {
			t.Error("Expecting dowloader been called")
		}
		checkFiles(t, []string{
			"data/Test",
			"data/Test/show.json",
			"data/Test/Season 01",
			"data/Test/Season 01/season.json",
			"data/Test/Season 01/Test s01e01 episode title.json",
			"data/Test/Season 01/Test s01e01 episode title.mp4"}, true)
		os.RemoveAll("data")
	})
	t.Run("Sad path", func(t *testing.T) {
		task := models.DownloadTask{Result: models.SearchResult{Title: "Test", Type: models.TypeSeries}}
		dl := newFakeDownloader(func(s *fakeDownloader) error { return os.ErrNotExist })
		c := makeMediaChannel(t, task, 1, dl)
		NewBatchDownloader(task.Result.Show, "data", models.DefaultFileNamer[models.PathNamingType(task.Result.Type)]).
			Download(context.
				Background(),
				c)

		if !dl.called {
			t.Error("Expecting dowloader been called")
		}
		checkFiles(t, []string{
			"data/Test",
			"data/Test/show.json",
			"data/Test/Season 01",
			"data/Test/Season 01/season.json",
			"data/Test/Season 01/Test s01e01 episode title.json",
			"data/Test/Season 01/Test s01e01 episode title.mp4"}, false)
		os.RemoveAll("data")
	})

	t.Run("Sad path one error", func(t *testing.T) {
		task := models.DownloadTask{Result: models.SearchResult{Title: "Test", Type: models.TypeSeries}}
		dl := newFakeDownloader(func(d *fakeDownloader) error {
			if d.count == 2 {
				return os.ErrInvalid
			}
			return nil
		})
		c := makeMediaChannel(t, task, 3, dl)
		NewBatchDownloader(task.Result.Show, "data", models.DefaultFileNamer[models.PathNamingType(task.Result.Type)]).
			Download(context.
				Background(),
				c)

		if !dl.called {
			t.Error("Expecting dowloader been called")
		}
		checkFiles(t, []string{
			"data/Test",
			"data/Test/show.json",
			"data/Test/Season 01",
			"data/Test/Season 01/season.json",
			"data/Test/Season 01/Test s01e01 episode title.json",
			"data/Test/Season 01/Test s01e01 episode title.mp4",
			"data/Test/Season 01/Test s01e03 episode title.json",
			"data/Test/Season 01/Test s01e03 episode title.mp4"}, true)
		checkFiles(t, []string{
			"data/Test/Season 01/Test s01e02 episode title.json",
			"data/Test/Season 01/Test s01e02 episode title.mp4",
		}, false)
		os.RemoveAll("data")
	})

}

func checkFiles(t *testing.T, names []string, exists bool) {
	t.Helper()
	for _, name := range names {
		if exist(name) != exists {
			t.Errorf("Expecting %q exist=%v", name, exists)
		}
	}
}

func exist(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

func makeMediaChannel(t *testing.T, task models.DownloadTask, number int, dl *fakeDownloader) <-chan models.DownloadItem {
	t.Helper()
	show := models.ShowInfo{
		ID:    uuid.NewString(),
		Plot:  "This is the show plot",
		Type:  task.Result.Type,
		Title: task.Result.Show,
		Images: []models.Image{
			{
				ID:     uuid.NewString(),
				Aspect: "banner",
				URL:    "https://via.placeholder.com/1920x1080.png",
			},
		},
	}

	season := models.SeasonInfo{
		ID:   uuid.NewString(),
		Plot: "This is the season plot",
		Images: []models.Image{
			{
				ID:     uuid.NewString(),
				Aspect: "banner",
				URL:    "https://via.placeholder.com/1920x1080.png",
			},
		},
	}
	c := make(chan models.DownloadItem, 1)

	go func() {
		defer close(c)
		for i := 0; i < number; i++ {
			c <- models.DownloadItem{
				Downloader: dl,
				MediaInfo: models.MediaInfo{
					ID:         uuid.NewString(),
					Show:       task.Result.Title,
					Title:      "episode title",
					Episode:    1 + i,
					Season:     1,
					Year:       1999,
					SeasonInfo: &season,
					Aired:      time.Now().AddDate(0, -1, i*7),
					Type:       task.Result.Type,
					ShowInfo:   &show,
					Plot:       "This is the episode plot",
				},
			}
		}
	}()

	return c
}

type fakeDownloader struct {
	called    bool
	mediafile string
	errFn     func(d *fakeDownloader) error
	count     int
}

func newFakeDownloader(errFn func(d *fakeDownloader) error) *fakeDownloader {
	return &fakeDownloader{
		errFn: errFn,
	}
}

func (d *fakeDownloader) Download(ctx context.Context, destination string) (err error) {
	d.called = true
	d.mediafile = destination
	f, _ := os.Create(destination)
	f.WriteString("Fake file")
	f.Close()
	d.count++
	return d.errFn(d)
}
func (d *fakeDownloader) WithProgresser(p download.Progresser)                     {}
func (d *fakeDownloader) WithLogger(l interface{ Printf(string, ...interface{}) }) {}
