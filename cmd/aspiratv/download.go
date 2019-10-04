package main

import (
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/simulot/aspiratv/metadata/nfo"

	"github.com/simulot/aspiratv/download"
	"github.com/simulot/aspiratv/providers"
	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
)

func (a *app) SubmitDownload(ctx context.Context, wg *sync.WaitGroup, p providers.Provider, s *providers.Show, d string, pc *mpb.Progress, bar *mpb.Bar) {
	wg.Add(1)
	go a.worker.Submit(func() {
		a.DownloadShow(ctx, p, s, d, pc)
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

func (a *app) DownloadShow(ctx context.Context, p providers.Provider, s *providers.Show, d string, pc *mpb.Progress) {
	if ctx.Err() != nil {
		return
	}
	id := 1000 + atomic.AddInt32(&dlID, 1)
	if a.Config.Debug {
		log.Printf("[%s] Starting  DownloadShow %d", p.Name(), id)
	}

	url, err := p.GetShowStreamURL(ctx, s)
	if err != nil {
		log.Println(err)
		return
	}
	if len(url) == 0 {
		log.Printf("[%s] Can't get url from %s.", p.Name(), providers.GetShowFileName(ctx, s))
		return
	}

	var pgr *progressBar
	if !a.Config.Headless {
		pgr = a.NewDownloadBar(pc, filepath.Base(providers.GetShowFileName(ctx, s)), id)
	}
	// Make a context for DownloadShow
	files := []string{}
	shouldDeleteFile := false

	fn := filepath.Join(d, providers.GetShowFileName(ctx, s))
	if a.Config.Debug {
		log.Printf("[%s] Downloading into file: %q", p.Name(), fn)
	}
	defer func() {
		if shouldDeleteFile {
			log.Printf("[%s] %s is cancelled.", p.Name(), providers.GetShowFileName(ctx, s))
			for _, f := range files {
				log.Printf("[%s] Remove %q.", p.Name(), f)
				err := os.Remove(f)
				if err != nil {
					log.Printf("[%s] Can't remove %q: %w.", p.Name(), f, err)
				}
			}
		}
		if a.Config.Debug {
			log.Printf("DownloadShow %d terminated", id)
		}
		if !a.Config.Headless {
			pgr.bar.SetTotal(1, true)
		}
	}()

	if a.Config.Debug {
		log.Printf("[%s] Download stream to: %q", p.Name(), fn)
	}
	err = os.MkdirAll(filepath.Dir(fn), 0777)
	if err != nil {
		log.Println(err)
		return
	}

	if a.Config.Debug {
		log.Println("Download url: ", url)
	}

	params := []string{
		"-loglevel", "info", // I wan't errors
		"-hide_banner", // I don't want banner
		"-i", url,      // Where is the stream
		"-metadata", "title=" + s.Title, // Force title
		"-metadata", "comment=" + s.Pitch, // Force comment
		"-metadata", "show=" + s.Show, //Force show
		"-metadata", "channel=" + s.Channel, // Force channel
		"-y",              // Override output file
		"-vcodec", "copy", // copy video
		"-acodec", "copy", // copy audio
		"-bsf:a", "aac_adtstoasc", // I don't know
		fn, // output file
	}

	if a.Config.Debug {
		log.Printf("[%s] Downloading %q", p.Name(), providers.GetShowFileName(ctx, s))
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

	// Then download thumbnail
	tbnFileName := strings.TrimSuffix(fn, filepath.Ext(fn)) + filepath.Ext(s.ThumbnailURL)
	showTbnFileName := filepath.Join(filepath.Dir(filepath.Dir(fn)), "show"+filepath.Ext(s.ThumbnailURL))
	mustDownloadShowTbnFile := false
	if _, err := os.Stat(showTbnFileName); os.IsNotExist(err) {
		mustDownloadShowTbnFile = true
	}

	tbnStream, err := a.getter.Get(ctx, s.ThumbnailURL)
	if err != nil {
		log.Printf("[%s] Can't download %q's thumbnail: %v", p.Name(), providers.GetShowFileName(ctx, s), err)
	}
	ws := []io.Writer{}
	tbnFile, err := os.Create(tbnFileName)
	if err != nil {
		log.Printf("[%s] Can't create %q's thumbnail: %v", p.Name(), providers.GetShowFileName(ctx, s), err)
	}
	defer tbnFile.Close()
	ws = append(ws, tbnFile)

	if mustDownloadShowTbnFile {
		showTbnFile, err := os.Create(showTbnFileName)
		if err != nil {
			log.Printf("[%s] Can't create shows's %q thumbnail: %v", p.Name(), s.Show, err)
		}
		defer showTbnFile.Close()
		ws = append(ws, showTbnFile)
	}

	wr := io.MultiWriter(ws...)
	_, err = io.Copy(wr, tbnStream)
	if err != nil {
		log.Printf("[%s] Can't write %q's thumbnail: %v", p.Name(), providers.GetShowFileName(ctx, s), err)
	}
	if a.Config.Headless || a.Config.Debug {
		log.Printf("[%s] %s downloaded.", p.Name(), providers.GetShowFileName(ctx, s))
	}
	return
}

func (a *app) CheckNFO(ctx context.Context, p providers.Provider, s *providers.Show, d string) {
	if s.IsSerie() {
		p := nfo.ShowNFOPath(s)
		_, err := os.Stat(p)
		if os.IsNotExist(err) {
			nfo.WriteShowData(s)
		}

		p = nfo.EpisodeNFOPath(s)
		_, err = os.Stat(p)
		if os.IsNotExist(err) {
			nfo.WriteEpisodeData(s)
		}
	}
}
