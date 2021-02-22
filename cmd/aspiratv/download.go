package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/simulot/aspiratv/download"
	"github.com/simulot/aspiratv/media"
	"github.com/simulot/aspiratv/metadata/nfo"
	"github.com/simulot/aspiratv/providers"
	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
)

type progressBar struct {
	lastSize int64
	start    time.Time
	bar      *mpb.Bar
}

// NewDownloadBar create a progress bar when not headless, nil otherwise.
// A nil progress bar is safe
func (a *app) NewDownloadBar(pc *mpb.Progress, name string, id int32) *progressBar {
	if a.Config.Headless {
		return nil
	}
	b := &progressBar{}
	b.bar = pc.AddBar(100*1024*1024*1024,
		mpb.BarWidth(12),
		mpb.AppendDecorators(
			decor.AverageSpeed(decor.UnitKB, " %.1f", decor.WC{W: 15, C: decor.DidentRight}),
			decor.Name(name),
		),
		mpb.BarRemoveOnComplete(),
	)
	b.bar.SetPriority(int(id))
	return b
}

// Init the progress bar
func (p *progressBar) Init(totalCount int64) {
	if p != nil {
		p.start = time.Now()
	}
}

// Update the progress bar,
func (p *progressBar) Update(count int64, size int64) {
	if p != nil && p.bar != nil {
		p.bar.SetTotal(size, count >= size)
		p.bar.IncrInt64(count-p.lastSize, time.Since(p.start))
		p.lastSize = count
	}
}

// Download count
var dlID = int32(0)

// DownloadShow does the actual download work
func (a *app) DownloadShow(ctx context.Context, p providers.Provider, m *media.Media, pc *mpb.Progress) {
	ctx, cancel := context.WithCancel(ctx)

	var dlErr error
	var itemName string
	showPath := m.ShowPath

	info := m.Metadata.GetMediaInfo()

	files := []string{} // Collect files beeing downloaded and to be deleted in case of cancellation

	defer func() {
		if dlErr != nil || ctx.Err() != nil {
			// on error, cleans up already present files
			a.logger.Info().Printf("[%s] download of %q is cancelled.", p.Name(), itemName)
			for _, f := range files {
				a.logger.Debug().Printf("[%s] Removing %q.", p.Name(), f)
				err := os.Remove(f)
				if err != nil {
					a.logger.Error().Printf("[%s] Error %s", p.Name(), err)
				}
			}
		}
		cancel()
	}()

	// One more download
	id := 1000 + atomic.AddInt32(&dlID, 1)

	mediaPath, err := download.MediaPath(showPath, m.Match, info)
	if err != nil {
		a.logger.Fatal().Printf("[%s] Can't determine file name.", p.Name(), err)
	}

	itemName = filepath.Base(mediaPath)

	err = os.MkdirAll(filepath.Dir(mediaPath), 0777)
	if err != nil {
		a.logger.Error().Printf("Can't create destination folder: %s", err)
		return
	}

	// Get more details from the provider
	err = p.GetMediaDetails(ctx, m)
	url := m.Metadata.GetMediaInfo().MediaURL
	if err != nil {
		a.logger.Error().Printf("[%s] Can't get url from %s: %s", p.Name(), itemName, err)
		return
	}
	if len(url) == 0 {
		a.logger.Error().Printf("[%s] Can't get url from %s: empty url", p.Name(), itemName)
		return
	}
	a.logger.Trace().Printf("[%s] Player URL for '%s' is %q ", p.Name(), itemName, url)
	a.logger.Trace().Printf("[%s] Start downloading media into %q", p.Name(), mediaPath)

	var pgr *progressBar
	if !a.Config.Headless {
		pgr = a.NewDownloadBar(pc, itemName, id)
		defer pgr.bar.SetTotal(1, true)
	}
	files = append(files, mediaPath) // remember it

	dlErr = download.Download(ctx, a.logger, url, mediaPath, info, download.WithProgress(pgr), download.WithLogger(a.logger))

	if dlErr != nil {
		a.logger.Error().Printf("[%s] Download exits with error:\n%s", p.Name(), dlErr)
		return
	}

	if a.Config.WriteNFO {
		a.WriteInfo(ctx, p, mediaPath, m, pc, id, &files)
		if ctx.Err() != nil {
			return
		}
	}
	if ctx.Err() != nil {
		a.logger.Error().Printf("[%s] Download exits with error:\n%s", p.Name(), dlErr)
		return
	}

	a.logger.Info().Printf("[%s] Media %q downloaded.", p.Name(), itemName)
	return
}

// WriteInfo writes nfo files attached to the media, and when present to the season and the show.
func (a *app) WriteInfo(ctx context.Context, p providers.Provider, showPath string, m *media.Media, pc *mpb.Progress, id int32, downloadedFiles *[]string) {
	var metaBar *mpb.Bar

	defer func() {
		if metaBar != nil {
			metaBar.SetTotal(1, true)
			metaBar = nil
		}
	}()

	info := m.Metadata.GetMediaInfo()
	mediaPath, err := download.MediaPath(m.ShowPath, m.Match, info)

	if !a.Config.Headless {
		metaBar = pc.AddBar(100*1024*1024*1024,
			mpb.BarWidth(12),
			mpb.AppendDecorators(
				decor.Name("Get images for", decor.WC{W: 15, C: decor.DidentRight}),
				decor.Name(filepath.Base(mediaPath)),
			),
			mpb.BarRemoveOnComplete(),
		)
		metaBar.SetPriority(int(id))
	}

	if info.MediaType != nfo.TypeMovie {
		if info.TVShow != nil {
			nfoPath := filepath.Join(showPath, "tvshow.nfo")
			nfoExists, err := fileExists(nfoPath)
			if !nfoExists && err == nil {
				info.TVShow.WriteNFO(nfoPath)
				if err != nil {
					a.logger.Error().Printf("WriteNFO: %s", err)
				}
				*downloadedFiles = append(*downloadedFiles, nfoPath)
				a.DownloadImages(ctx, p, nfoPath, info.TVShow.Thumb, downloadedFiles)
			}
		}
		if info.SeasonInfo != nil {
			seasonPath, err := download.SeasonPath(showPath, m.Match, info)
			if err != nil {
				a.logger.Error().Printf("Can't dertermine season path: %s", err)
			}
			nfoPath := filepath.Join(seasonPath, "season.nfo")
			nfoExists, err := fileExists(nfoPath)
			if !nfoExists && err == nil {
				info.SeasonInfo.WriteNFO(nfoPath)
				if err != nil {
					a.logger.Error().Printf("WriteNFO: %s", err)
				}
				*downloadedFiles = append(*downloadedFiles, nfoPath)
				a.DownloadImages(ctx, p, nfoPath, info.SeasonInfo.Thumb, downloadedFiles)
			}
		}
	}

	nfoPath := strings.TrimSuffix(mediaPath, filepath.Ext(mediaPath)) + ".nfo"
	nfoExists, err := fileExists(nfoPath)
	if !nfoExists && err == nil {
		err = m.Metadata.WriteNFO(nfoPath)
		if err != nil {
			a.logger.Error().Printf("WriteNFO: %s", err)
		}
		*downloadedFiles = append(*downloadedFiles, nfoPath)
		a.DownloadImages(ctx, p, nfoPath, info.Thumb, downloadedFiles)
	}

}

// DownloadImages images listed in nfo.Thumbs into nfoPath. nfo's name  indicates if images are those from episode, season or show.
func (a *app) DownloadImages(ctx context.Context, p providers.Provider, nfoPath string, thumbs []nfo.Thumb, downloadedFiles *[]string) {

	nfoDir := filepath.Dir(nfoPath)
	nfoName := filepath.Base(nfoPath)
	err := os.MkdirAll(nfoDir, 0777)
	if err != nil {
		a.logger.Error().Printf("[%s] Can't create dir %s", p.Name(), err)
		return
	}

	for i, thumb := range thumbs {
		var base string
		if ctx.Err() != nil {
			a.logger.Debug().Printf("[%s] Cancelling %s", p.Name(), ctx.Err())
			return
		}

		switch nfoName {
		case "tvshow.nfo", "season.nfo":
			base = fmt.Sprintf("%s_%d.png", thumb.Aspect, i+1)
		default: // For episodes
			if thumb.Aspect == "thumb" || thumb.Aspect == "" {
				base = fmt.Sprintf("%s_%d.png", nfoName, i+1)
			} else {
				continue
			}
		}
		thumbName := filepath.Join(nfoDir, base)
		if thumbExists, _ := fileExists(thumbName); thumbExists {
			continue
		}
		err := a.DownloadImage(ctx, thumb.URL, thumbName, downloadedFiles)

		if err != nil {
			a.logger.Error().Printf("[%s] Can't get thumbnail from %q: %s", p.Name(), thumb.URL, err)
			continue
		}
		a.logger.Debug().Printf("[%s] thumbnail %q downloaded.", p.Name(), thumbName)
	}
}
