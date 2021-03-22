package providers

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/simulot/aspiratv/download"
	"github.com/simulot/aspiratv/matcher"
	"github.com/simulot/aspiratv/media"
	"github.com/simulot/aspiratv/mylog"
	"github.com/simulot/aspiratv/workers"
)

func ErrString(err error) interface{} {
	if err == nil {
		return "NoErr"
	}
	return err
}

type FeedBacker interface {
	Stage(stage string) // Indicate current stage
	Total(total int)    // Indicate the total number (could be bytes, percent )
	Update(current int) // Indicate the current position
	Done()              // Call when the task is done
}

type RunnerConfig struct {
	log                *mylog.MyLog
	concurentDownloads int
}

type RunnerConfigFn func(c RunnerConfig) RunnerConfig

// Runner handle downloads for a provider
type Runner struct {
	p  Provider
	s  *Settings
	w  *workers.Worker
	c  RunnerConfig
	wg sync.WaitGroup
}

// NewRunner return  a configured runner for the provider
func NewRunner(ctx context.Context, s *Settings, p Provider, fns ...RunnerConfigFn) *Runner {
	r := &Runner{
		p: p,
		s: s,
	}

	c := RunnerConfig{
		concurentDownloads: runtime.NumCPU(),
	}
	for _, f := range fns {
		c = f(c)
	}
	r.c = c

	r.w = workers.New(ctx, c.concurentDownloads, r.c.log)
	return r
}

// GetNewMediasList pull the providers for medias available on its web site, check if
// it isn't already downloaded before sent it back to the channel
func (r *Runner) GetNewMediasList(ctx context.Context, mr []*matcher.MatchRequest) <-chan *media.Media {
	r.c.log.Trace().Printf("[%s] Runner.GetAvailableMedias", r.p.Name())
	c := make(chan *media.Media)
	r.wg.Add(1)
	go func() {
		defer func() {
			close(c)
			r.c.log.Trace().Printf("[%s] Exit Runner.GetAvailableMedias, %s", r.p.Name(), ErrString(ctx.Err()))
			r.wg.Done()
		}()
		seen := map[string]bool{}

		for m := range r.p.MediaList(ctx, mr) {
			if _, ok := seen[m.ID]; ok {
				// Skip seen media
				continue
			}
			seen[m.ID] = true

			m.ShowRootPath = m.Match.ShowRootPath
			if len(m.ShowRootPath) == 0 {
				m.ShowRootPath = filepath.Join(r.s.Destinations[m.Match.Destination], download.PathNameCleaner(m.Metadata.GetMediaInfo().Showtitle))
			}
			showPath, err := download.MediaPath(m.ShowRootPath, m.Match, m.Metadata.GetMediaInfo())
			if err != nil {
				r.c.log.Error().Printf("[%s] %s", r.p.Name())
				continue
			}
			exist, err := fileExists(showPath)
			if err != nil {
				r.c.log.Error().Printf("[%s] %s", r.p.Name())
				continue
			}
			if exist {
				r.c.log.Trace().Printf("[%s] Media already downloaded %q", r.p.Name(), showPath)
				continue
			}
			select {
			case <-ctx.Done():
				return
			default:
				c <- m
			}
		}
	}()
	return c
}

func (r *Runner) WaitUntilCompletion(ctx context.Context) {
	r.w.Stop(ctx)
	r.wg.Wait()
	r.c.log.Trace().Printf("[runner] [%s] Runner.WaitUntilCompletion done (%s)", r.p.Name(), ErrString(ctx.Err()))
}

func (r *Runner) SubmitDownload(ctx context.Context, m *media.Media, fb FeedBacker, onDone func()) {
	r.c.log.Trace().Printf("[runner] [%s] Runner.Enqueue download of %q", r.p.Name(), m.Metadata.GetMediaInfo().Title)
	r.w.Submit(ctx, func(wid, jid int) {
		r.c.log.Trace().Printf("[runner] [%s] Job (%d,%d) Started  %q", r.p.Name(), wid, jid, m.Metadata.GetMediaInfo().Title)
		defer func() {
			if onDone != nil {
				r.c.log.Trace().Printf("[runner] [%s] Job (%d,%d) Done function called for %s", r.p.Name(), wid, jid, m.Metadata.GetMediaInfo().Title)
				onDone()
			}
			r.c.log.Trace().Printf("[runner] [%s] Job (%d,%d) is done  of %q (%s)", r.p.Name(), wid, jid, m.Metadata.GetMediaInfo().Title, ErrString(ctx.Err()))
		}()
		newDownloader(r).download(ctx, m, fb)

	})
}

func RunnerWithLogger(log *mylog.MyLog) RunnerConfigFn {
	return func(c RunnerConfig) RunnerConfig {
		c.log = log
		return c
	}
}

func RunnerWithConcurentLimit(limit int) RunnerConfigFn {
	return func(c RunnerConfig) RunnerConfig {
		c.concurentDownloads = limit
		return c
	}
}

func fileExists(p string) (bool, error) {
	_, err := os.Stat(p)
	if err != nil {
		if os.IsExist(err) {
			return true, nil
		}
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
