package main

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/simulot/aspiratv/download"
	"github.com/simulot/aspiratv/matcher"
	"github.com/simulot/aspiratv/providers"
	"golang.org/x/time/rate"
)

// Run command

func (a *app) Run(ctx context.Context) {
	a.logger.Trace().Printf("[RUN] Start RUN command")
	defer func() {
		a.logger.Trace().Printf("[RUN] Exit RUN command")
	}()
	if !a.Headless {
		a.BarContainer = NewBarContainer(ctx)
		defer func() {
			a.BarContainer.Done()
		}()
	}

	err := a.ReadConfig(a.ConfigFile)
	if err != nil {
		a.logger.Fatal().Printf("[Initialize] %s", err)
	}
	wg := sync.WaitGroup{}
	idx := 0
	for _, p := range providers.List() {
		if !a.Settings.Providers[p.Name()].Enabled {
			continue
		}
		mrs := []*matcher.MatchRequest{}
		for _, mr := range a.Settings.WatchList {
			if mr.Provider == p.Name() {
				mrs = append(mrs, mr)
			}
		}
		if len(mrs) > 0 {
			wg.Add(1)
			go func(p providers.Provider, mrs []*matcher.MatchRequest, idx int) {
				defer func() {
					wg.Done()
				}()
				a.GetMediasOfProvider(ctx, p, idx, mrs)
			}(p, mrs, idx)
			idx++
		}
	}
	wg.Wait()
}

type nextID int64

func (n *nextID) Next() int { return int(atomic.AddInt64((*int64)(n), 1)) }

func (a *app) GetMediasOfProvider(ctx context.Context, p providers.Provider, idx int, mrs []*matcher.MatchRequest) {
	a.logger.Trace().Printf("[RUN] Start GetMediasOfProvider(%s)", p.Name())
	defer func() {
		a.logger.Trace().Printf("[RUN] Exit GetMediasOfProvider(%s)", p.Name())
	}()
	var pBar *bar
	next := nextID(idx * 10000)

	if !a.Headless {
		pBar = a.BarContainer.NewProviderBar(p.Name(), idx)
		defer func() {
			pBar.Done()
		}()
	}

	hitsPerSecond := a.Settings.Providers[p.Name()].HitsRate
	if hitsPerSecond == 0 {
		hitsPerSecond = 5
	}
	mediaCount := 0
	mediaDone := 0
	p.Configure(
		providers.ProviderLog(a.logger),
		providers.ProviderHitsPerSecond(rate.NewLimiter(rate.Limit(hitsPerSecond), 2*hitsPerSecond)),
	)

	r := providers.NewRunner(ctx, &a.Settings, p,
		providers.RunnerWithLogger(a.logger),
		providers.RunnerWithConcurentLimit(a.ConcurrentTasks),
	)
	defer func() {
		a.logger.Trace().Printf("[RUN] GetMediasOfProvider(%s): WaitUntilCompletion", p.Name())
		r.WaitUntilCompletion(ctx)
		a.logger.Trace().Printf("[RUN] GetMediasOfProvider(%s): completed", p.Name())
	}()

	for m := range r.GetNewMediasList(ctx, mrs) {
		select {
		case <-ctx.Done():
			a.logger.Trace().Printf("[RUN] GetMediasOfProvider(%s): Cancellation received", p.Name())
			return
		default:
			mediaCount++
			if !a.Headless {
				pBar.Total(mediaCount)
			}
			m := m
			mediaPath, _ := download.MediaPath(m.ShowRootPath, m.Match, m.Metadata.GetMediaInfo())
			dlBar := a.BarContainer.NewDownloadBar(filepath.Base(mediaPath), next.Next())
			r.SubmitDownload(ctx, m, dlBar, func() {

				mediaDone++
				if !a.Headless {
					pBar.Update(mediaDone)
					dlBar.Done()
				} else {
					fmt.Printf("[%s] File %q downloaded\n", p.Name(), mediaPath)

				}
			})
		}
	}
}
