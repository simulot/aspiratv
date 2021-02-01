package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/simulot/aspiratv/download"
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

func (p *progressBar) Init(totalCount int64) {
	if p != nil {
		p.start = time.Now()
	}
}

func (p *progressBar) Update(count int64, size int64) {
	if p != nil && p.bar != nil {
		p.bar.SetTotal(size, count >= size)
		p.bar.IncrInt64(count-p.lastSize, time.Since(p.start))
		p.lastSize = count
	}
}

var dlID = int32(0)

func (a *app) DownloadShow(ctx context.Context, p providers.Provider, m *providers.Media, pc *mpb.Progress) {
	ctx, cancel := context.WithCancel(ctx)

	var dlErr error
	var itemName string
	// Collect files beeing downloaded and to be deleted in case of cancellation
	files := []string{}

	defer func() {
		if dlErr != nil || ctx.Err() != nil {
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
	id := 1000 + atomic.AddInt32(&dlID, 1)
	ShowPath := m.ShowPath

	itemName = filepath.Base(m.Metadata.GetMediaPath(ShowPath))

	err := p.GetMediaDetails(ctx, m) // Side effect: Episode number can be determined at this point.
	url := m.Metadata.GetMediaInfo().MediaURL
	if err != nil || len(url) == 0 {
		a.logger.Error().Printf("[%s] Can't get url from %s.", p.Name(), filepath.Base(m.Metadata.GetMediaPath(ShowPath)))
		return
	}

	if a.Config.WriteNFO {
		a.DownloadInfo(ctx, p, ShowPath, m, pc, id, &files)
		if ctx.Err() != nil {
			return
		}
	}

	var pgr *progressBar
	fn := m.Metadata.GetMediaPath(ShowPath)
	itemName = filepath.Base(fn)

	a.logger.Trace().Printf("[%s] Start downloading media %q", p.Name(), fn)

	if !a.Config.Headless {
		pgr = a.NewDownloadBar(pc, filepath.Base(fn), id)
		defer pgr.bar.SetTotal(1, true)
	}
	a.logger.Debug().Printf("[%s] Stream url: %q", p.Name(), url)

	err = os.MkdirAll(filepath.Dir(fn), 0777)
	if err != nil {
		a.logger.Error().Printf("Can't create destination folder: %s", err)
		return
	}

	info := m.Metadata.GetMediaInfo()

	files = append(files, fn)
	dlErr = download.Download(ctx, a.logger, url, fn, info, download.WithProgress(pgr))

	if dlErr != nil {
		a.logger.Error().Printf("[%s] Download exits with error:\n%s", p.Name(), dlErr)
		return
	}

	if ctx.Err() != nil {
		a.logger.Error().Printf("[%s] Download exits with error:\n%s", p.Name(), dlErr)
		return
	}

	a.logger.Info().Printf("[%s] Media %q downloaded.", p.Name(), filepath.Base(fn))
	return
}

func (a *app) DownloadInfo(ctx context.Context, p providers.Provider, destination string, m *providers.Media, pc *mpb.Progress, id int32, downloadedFiles *[]string) {

	var metaBar *mpb.Bar
	itemName := filepath.Base(m.Metadata.GetMediaPath(destination))

	defer func() {
		if metaBar != nil {
			metaBar.SetTotal(1, true)
			metaBar = nil
		}
	}()

	if !a.Config.Headless {
		metaBar = pc.AddBar(100*1024*1024*1024,
			mpb.BarWidth(12),
			mpb.AppendDecorators(
				decor.Name("Get images", decor.WC{W: 15, C: decor.DidentRight}),
				decor.Name(filepath.Base(itemName)),
			),
			mpb.BarRemoveOnComplete(),
		)
		metaBar.SetPriority(int(id))
	}

	info := m.Metadata.GetMediaInfo()
	nfoPath := m.Metadata.GetNFOPath(destination)
	nfoExists, err := fileExists(nfoPath)
	if !nfoExists && err == nil {
		err = m.Metadata.WriteNFO(nfoPath)
		if err != nil {
			a.logger.Error().Printf("WriteNFO: %s", err)
		}
		*downloadedFiles = append(*downloadedFiles, nfoPath)
		a.DowloadImages(ctx, p, nfoPath, info.Thumb, downloadedFiles)
	}
	if m.ShowType == providers.Series {
		seasonPath := filepath.Dir(m.Metadata.GetMediaPath(destination))
		if info.SeasonInfo != nil {
			nfoPath = filepath.Join(seasonPath, "season.nfo")
			nfoExists, err = fileExists(nfoPath)
			if !nfoExists && err == nil {
				info.SeasonInfo.WriteNFO(nfoPath)
				if err != nil {
					a.logger.Error().Printf("WriteNFO: %s", err)
				}
				*downloadedFiles = append(*downloadedFiles, nfoPath)

				a.DowloadImages(ctx, p, nfoPath, info.SeasonInfo.Thumb, downloadedFiles)
			}
		}
		if info.TVShow != nil {
			nfoPath = filepath.Join(destination, "tvshow.nfo")
			nfoExists, err = fileExists(nfoPath)
			if !nfoExists && err == nil {
				info.TVShow.WriteNFO(nfoPath)
				if err != nil {
					a.logger.Error().Printf("WriteNFO: %s", err)
				}
				*downloadedFiles = append(*downloadedFiles, nfoPath)

				a.DowloadImages(ctx, p, nfoPath, info.TVShow.Thumb, downloadedFiles)
			}
		}
	}
}

func (a *app) DowloadImages(ctx context.Context, p providers.Provider, destination string, thumbs []nfo.Thumb, downloadedFiles *[]string) {
	nfoFile := filepath.Base(destination)
	if filepath.Ext(destination) != "" {
		destination = filepath.Dir(destination)
	}

	err := os.MkdirAll(filepath.Dir(destination), 0777)
	if err != nil {
		a.logger.Error().Printf("[%s] Can't create %s", p.Name(), err)
		return
	}

	for _, thumb := range thumbs {
		var base string
		if ctx.Err() != nil {
			a.logger.Debug().Printf("[%s] Cancelling %s", p.Name(), ctx.Err())
			return
		}

		switch nfoFile {
		case "tvshow.nfo", "season.nfo":
			base = thumb.Aspect + ".png"
		default: // For episodes
			if thumb.Aspect == "thumb" {
				base = strings.TrimSuffix(nfoFile, filepath.Ext(nfoFile)) + ".png"
			} else {
				continue
			}
		}
		thumbName := filepath.Join(destination, base)
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
