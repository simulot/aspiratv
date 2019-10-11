package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/simulot/aspiratv/download"
	"github.com/simulot/aspiratv/metadata/nfo"
	"github.com/simulot/aspiratv/providers"
	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
)

func (a *app) SubmitDownload(ctx context.Context, wg *sync.WaitGroup, p providers.Provider, m *providers.Media, pc *mpb.Progress, bar *mpb.Bar) {
	wg.Add(1)
	go a.worker.Submit(func() {
		a.DownloadShow(ctx, p, m, pc)
		if bar != nil {
			bar.Increment()
		}
	}, wg)
}

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
	wg := sync.WaitGroup{}
	if ctx.Err() != nil {
		return
	}

	id := 1000 + atomic.AddInt32(&dlID, 1)

	err := p.GetMediaDetails(ctx, m) // Side effect: Episode number can be determined at this point.
	url := m.Metadata.GetMediaInfo().URL
	if err != nil || len(url) == 0 {
		log.Printf("[%s] Can't get url from %s.", p.Name(), filepath.Base(m.Metadata.GetMediaPath(a.Config.Destinations[m.Match.Destination])))
		return
	}

	if a.Config.WriteNFO {
		info := m.Metadata.GetMediaInfo()
		nfoPath := m.Metadata.GetNFOPath(a.Config.Destinations[m.Match.Destination])
		nfoExists, err := fileExists(nfoPath)
		if !nfoExists && err == nil {
			err = m.Metadata.WriteNFO(nfoPath)
			if err != nil {
				log.Println(err)
			}
			wg.Add(1)
			// go func() {
			a.DowloadImages(ctx, p, nfoPath, info.Thumb)
			wg.Done()
			// }()
		}
		if m.ShowType == providers.Series {
			if info.SeasonInfo != nil {
				nfoPath = m.Metadata.GetSeasonNFOPath(a.Config.Destinations[m.Match.Destination])
				nfoExists, err = fileExists(nfoPath)
				if !nfoExists && err == nil {
					info.SeasonInfo.WriteNFO(nfoPath)
					if err != nil {
						log.Println(err)
					}
					wg.Add(1)
					// go func() {
					a.DowloadImages(ctx, p, nfoPath, info.SeasonInfo.Thumb)
					wg.Done()
					// }()
				}
			}
			if info.TVShow != nil {
				nfoPath = m.Metadata.GetShowNFOPath(a.Config.Destinations[m.Match.Destination])
				nfoExists, err = fileExists(nfoPath)
				if !nfoExists && err == nil {
					info.TVShow.WriteNFO(nfoPath)
					if err != nil {
						log.Println(err)
					}
					wg.Add(1)
					// go func() {
					a.DowloadImages(ctx, p, nfoPath, info.TVShow.Thumb)
					wg.Done()
					// }()
				}
			}
		}
	}

	fn := m.Metadata.GetMediaPath(a.Config.Destinations[m.Match.Destination])
	if a.Config.Headless || a.Config.Debug {
		log.Printf("[%s] Start downloading media %q", p.Name(), fn)
	}

	if a.Config.Debug {
		log.Printf("[%s] Stream url: %q", p.Name(), url)
	}

	var pgr *progressBar
	if !a.Config.Headless {
		pgr = a.NewDownloadBar(pc, filepath.Base(fn), id)
	}
	// Make a context for DownloadShow
	files := []string{}
	shouldDeleteFile := false

	if a.Config.Debug {
		log.Printf("[%s] Downloading into file: %q", p.Name(), fn)
	}
	defer func() {
		if shouldDeleteFile {
			log.Printf("[%s] Cancelling download of %q.", p.Name(), filepath.Base(fn))
			for _, f := range files {
				log.Printf("[%s] Removing %q.", p.Name(), f)
				err := os.Remove(f)
				if err != nil {
					log.Printf("[%s] Can't remove %q: %s.", p.Name(), f, err)
				}
			}
		}
		if !shouldDeleteFile && (a.Config.Headless || a.Config.Debug) {
			log.Printf("[%s] Media %q downloaded", p.Name(), filepath.Base(fn))
		}
		if !a.Config.Headless {
			pgr.bar.SetTotal(1, true)
		}
	}()

	err = os.MkdirAll(filepath.Dir(fn), 0777)
	if err != nil {
		log.Println(err)
		return
	}

	info := m.Metadata.GetMediaInfo()

	params := []string{
		"-loglevel", "info", // Give me feedback
		"-hide_banner", // I don't want banner
		"-i", url,      // Where is the stream
		"-metadata", "title=" + info.Title, // Force title
		"-metadata", "comment=" + info.Plot, // Force comment
		"-metadata", "show=" + info.Showtitle, //Force show
		"-metadata", "channel=" + info.Studio, // Force channel
		"-y",              // Override output file
		"-vcodec", "copy", // copy video
		"-acodec", "copy", // copy audio
		"-bsf:a", "aac_adtstoasc", // I don't know
		fn, // output file
	}

	if a.Config.Debug {
		log.Printf("[%s] FFMPEG started %q", p.Name(), filepath.Base(fn))
	}

	files = append(files, fn)
	err = download.FFMepg(ctx, url, params, download.FFMepgWithProgress(pgr), download.FFMepgWithDebug(a.Config.Debug))

	if err != nil || ctx.Err() != nil {
		log.Printf("[%s] FFMEPG exits with error:\n%s", p.Name(), err)
		shouldDeleteFile = true
		return
	}

	if ctx.Err() != nil {
		shouldDeleteFile = true
		return
	}
	wg.Wait()
	if a.Config.Headless || a.Config.Debug {
		log.Printf("[%s] %q downloaded.", p.Name(), filepath.Base(fn))
	}
	return
}

func (a *app) DowloadImages(ctx context.Context, p providers.Provider, destination string, thumbs []nfo.Thumb) {

	nfoFile := filepath.Base(destination)
	if filepath.Ext(destination) != "" {
		destination = filepath.Dir(destination)
	}

	err := os.MkdirAll(filepath.Dir(destination), 0777)
	if err != nil {
		log.Printf("[%s] Can't create %s :%s", destination, err)
		return
	}

	for _, thumb := range thumbs {
		var base string

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
		err := a.DownloadToPNG(ctx, thumb.URL, thumbName)

		if err != nil {
			log.Printf("[%s] Can't get thumbnail from %q: %s", p.Name(), thumb.URL, err)
			continue
		}
		if a.Config.Headless || a.Config.Debug {
			log.Printf("[%s] thumbnail %q downloaded.", p.Name(), thumbName)
		}
	}
}
