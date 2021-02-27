package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/simulot/aspiratv/matcher"
	"github.com/simulot/aspiratv/providers"
)

func (a *app) Download(ctx context.Context, show string) {

	if a.Matcher.Provider == "" {
		a.Exit("Missing --provider")
	}

	if _, ok := providers.List()[a.Matcher.Provider]; !ok {
		a.Exit(fmt.Sprintf("Provider %q is unknown", a.Matcher.Provider))
	}

	if a.Matcher.ShowRootPath == "" {
		a.Exit("Missing --show-path")
	}

	v, err := providers.ExpandPath(a.Matcher.ShowRootPath)
	if err != nil {
		a.Exit("Invalid --show-path " + a.Matcher.ShowRootPath)
	}
	a.Matcher.ShowRootPath = v
	a.Matcher.Show = strings.ToLower(show)

	p, ok := providers.List()[a.Matcher.Provider]
	if !ok {
		a.logger.Fatal().Printf("Unknown provider %q", a.Matcher.Provider)
		os.Exit(1)
	}

	if !a.Headless {
		a.BarContainer = NewBarContainer(ctx)
		defer func() {
			a.BarContainer.Done()
		}()
	}

	mrs := make([]*matcher.MatchRequest, 0)
	mrs = append(mrs, &a.Matcher)

	if len(mrs) > 0 {
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func(p providers.Provider, mrs []*matcher.MatchRequest) {
			defer func() {
				wg.Done()
			}()
			a.GetMediasOfProvider(ctx, p, 0, mrs)
		}(p, mrs)
		wg.Wait()
	}
}
