package download

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestFFMPEG(t *testing.T) {
	t.Run("test inexistant url", func(t *testing.T) {
		os.RemoveAll("data")
		d := NewFFMPEG().Input("https://inexistant.domain.com/")
		err := d.Download(context.TODO(), "data/test.mp4")
		t.Logf("Error was: %s", err)
		if err == nil {
			t.Errorf("Error was expected")
		}
	})
	t.Run("test wrong output file", func(t *testing.T) {
		os.RemoveAll("data")
		d := NewFFMPEG().Input("http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4")
		err := d.Download(
			context.TODO(),
			"data/test.mp4")
		t.Logf("Error was: %s", err)
		if err == nil {
			t.Errorf("Error was expected")
		}
	})

	t.Run("interrupted", func(t *testing.T) {
		os.RemoveAll("data")
		os.Mkdir("data", 0777)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		d := NewFFMPEG().Input("http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4")
		time.AfterFunc(100*time.Millisecond, func() {
			cancel()
		})
		err := d.Download(
			ctx,
			"data/test.mp4")
		if err == nil {
			t.Errorf("Error was expected")
		}
	})

	t.Run("happy test", func(t *testing.T) {

		os.RemoveAll("data")
		os.Mkdir("data", 0777)
		p := pgr{t: t}
		d := NewFFMPEG().
			Input("http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4")
		d.WithProgresser(&p)

		err := d.Download(
			context.TODO(),
			"data/test.mp4")
		if err != nil {
			t.Errorf("Error was unexpected: %s", err)
		}
		if !p.called {
			t.Error("Progresser not called")
		}
	})

}

type pgr struct {
	t      *testing.T
	value  []struct{ current, total int }
	called bool
}

func (p *pgr) Progress(current int, total int) {
	p.value = append(p.value, struct{ current, total int }{current, total})
	p.t.Logf("progress: %d, %d", current, total)
	p.called = true
}
