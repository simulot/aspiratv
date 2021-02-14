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

func (a *app) DownloadShow(ctx context.Context, p providers.Provider, m *media.Media, pc *mpb.Progress) {
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
	showPath := m.ShowPath
	mediaPath, err := download.MediaPath(showPath, m.Match, m.Metadata.GetMediaInfo())
	if err != nil {
		a.logger.Fatal().Printf("[%s] Can't determine file name.", p.Name(), err)
	}

	itemName = filepath.Base(mediaPath)
	d := filepath.Dir(showPath)
	err = os.MkdirAll(d, 0777)
	if err != nil {
		a.logger.Error().Printf("Can't create destination folder: %s", err)
		return
	}

	err = p.GetMediaDetails(ctx, m)
	url := m.Metadata.GetMediaInfo().MediaURL
	if err != nil || len(url) == 0 {
		a.logger.Error().Printf("[%s] Can't get url from %s.", p.Name(), itemName)
		return
	}

	if a.Config.WriteNFO {
		a.DownloadInfo(ctx, p, showPath, m, pc, id, &files)
		if ctx.Err() != nil {
			return
		}
	}

	var pgr *progressBar

	a.logger.Trace().Printf("[%s] Start downloading media %q", p.Name(), showPath)

	if !a.Config.Headless {
		pgr = a.NewDownloadBar(pc, itemName, id)
		defer pgr.bar.SetTotal(1, true)
	}
	a.logger.Debug().Printf("[%s] Stream url: %q", p.Name(), url)

	info := m.Metadata.GetMediaInfo()

	files = append(files, showPath)
	dlErr = download.Download(ctx, a.logger, url, mediaPath, info, download.WithProgress(pgr), download.WithLogger(a.logger))

	if dlErr != nil {
		a.logger.Error().Printf("[%s] Download exits with error:\n%s", p.Name(), dlErr)
		return
	}

	if ctx.Err() != nil {
		a.logger.Error().Printf("[%s] Download exits with error:\n%s", p.Name(), dlErr)
		return
	}

	a.logger.Info().Printf("[%s] Media %q downloaded.", p.Name(), itemName)
	return
}

func (a *app) DownloadInfo(ctx context.Context, p providers.Provider, showPath string, m *media.Media, pc *mpb.Progress, id int32, downloadedFiles *[]string) {

	var metaBar *mpb.Bar
	itemName := filepath.Base(showPath)

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
				decor.Name("Get images for", decor.WC{W: 15, C: decor.DidentRight}),
				decor.Name(filepath.Base(itemName)),
			),
			mpb.BarRemoveOnComplete(),
		)
		metaBar.SetPriority(int(id))
	}

	info := m.Metadata.GetMediaInfo()
	mediaPath, err := download.MediaPath(showPath, m.Match, info)
	nfoPath := strings.TrimSuffix(mediaPath, filepath.Ext(mediaPath)) + ".nfo"
	nfoExists, err := fileExists(nfoPath)
	if !nfoExists && err == nil {
		err = m.Metadata.WriteNFO(nfoPath)
		if err != nil {
			a.logger.Error().Printf("WriteNFO: %s", err)
		}
		*downloadedFiles = append(*downloadedFiles, nfoPath)
		a.DowloadImages(ctx, p, nfoPath, info.Thumb, downloadedFiles)
	}
	if info.MediaType != nfo.TypeMovie {
		if info.SeasonInfo != nil {
			seasonPath, err := download.SeasonPath(showPath, m.Match, info)
			if err != nil {
				a.logger.Error().Printf("Can't dertermine season path: %s", err)
			}
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
			nfoPath = filepath.Join(showPath, "tvshow.nfo")
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

func (a *app) DowloadImages(ctx context.Context, p providers.Provider, nfoPath string, thumbs []nfo.Thumb, downloadedFiles *[]string) {
	nfoDir := filepath.Dir(nfoPath)
	nfoName := filepath.Base(nfoPath)
	nfoName = strings.TrimSuffix(nfoName, filepath.Ext(nfoName))
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
		case "tvshow", "season":
			base = fmt.Sprintf("%s_%d.png", thumb.Aspect, i+1)
		default: // For episodes
			if thumb.Aspect == "thumb" {
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
