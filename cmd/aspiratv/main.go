package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/simulot/aspiratv/providers/artetv"
	_ "github.com/simulot/aspiratv/providers/francetv"
	_ "github.com/simulot/aspiratv/providers/gulli"

	"github.com/simulot/aspiratv/net/http"
	"github.com/simulot/aspiratv/playlists/m3u8"
	"github.com/simulot/aspiratv/providers"

	"github.com/simulot/aspiratv/workers"
	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {

	fmt.Printf("%s: %v, commit %v, built at %v\n", filepath.Base(os.Args[0]), version, commit, date)
	a := &app{
		Stop: make(chan bool),
	}

	// trap Ctrl+C and call cancel on the context
	ctx, cancel := context.WithCancel(context.Background())
	breakChannel := make(chan os.Signal, 1)
	signal.Notify(breakChannel, os.Interrupt)

	defer func() {
		// Normal end... cleaning up
		signal.Stop(breakChannel)
		cancel()
	}()

	// waiting for interruption
	go func() {
		select {
		case <-breakChannel:
			cancel()
		case <-ctx.Done():
			return
		}
	}()

	cliConfig := &Config{}

	flag.BoolVar(&cliConfig.Debug, "debug", false, "Debug mode.")
	flag.BoolVar(&cliConfig.Force, "force", false, "Force media download.")
	flag.BoolVar(&cliConfig.Headless, "headless", false, "Headless mode. Progression bars are not displayed.")
	flag.StringVar(&cliConfig.ConfigFile, "config", "config.json", "Configuration file name.")
	flag.IntVar(&cliConfig.ConcurrentTasks, "max-tasks", runtime.NumCPU(), "Maximum concurrent downloads at a time.")
	flag.Parse()

	log.SetOutput(os.Stderr)

	a.Initialize(cliConfig)
	a.Run(ctx)
}

func (a *app) Initialize(c *Config) {
	a.Config = ReadConfigOrDie(c)

	// Check ans normalize configuration file
	a.Config.Check()

	// Check ffmpeg presence
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("where", "ffmpeg")
	} else {
		cmd = exec.Command("which", "ffmpeg")
	}
	b, err := cmd.Output()
	if err != nil {
		log.Fatal("Missing ffmpeg on your system, it's required to download video files.")
	}
	a.ffmpeg = strings.Trim(strings.Trim(string(b), "\r\n"), "\n")
	if a.Config.Debug {
		log.Printf("FFMPEG path: %q", a.ffmpeg)
	}
}

func (a *app) Run(ctx context.Context) {

	a.worker = workers.New(ctx, a.Config.ConcurrentTasks, a.Config.Debug)
	a.getter = http.DefaultClient

	var pc *mpb.Progress

	if !a.Config.Headless {
		pc = mpb.NewWithContext(
			ctx,
			mpb.WithWidth(64),
			mpb.ContainerOptOnCond(
				mpb.WithOutput(nil),
				func() bool {
					return a.Config.Headless
				},
			))
	}

	activeProviders := int64(0)
	for _, p := range providers.List() {
		if a.Config.IsProviderActive(p.Name()) {
			activeProviders++
		}
	}

providerLoop:
	for _, p := range providers.List() {
		if a.Config.IsProviderActive(p.Name()) {
			select {
			case <-ctx.Done():
				break providerLoop
			default:
				if ctx.Err() != nil {
					break providerLoop
				}
				a.PullShows(ctx, p, pc)
			}
		}
	}
	if !a.Config.Headless {
		pc.Wait()
	}
	if a.Config.Debug {
		log.Println("End of providerLoop")
	}
	a.worker.Stop()
	if a.Config.Debug {
		log.Println("Workers stop confirmed")
	}
	if a.Config.Debug {
		log.Println("End of Run")
	}
}

type debugger interface {
	SetDebug(bool)
}

func left(s string, l int) string {
	if len(s) > l {
		return s[:l]
	}
	return s
}

// PullShows pull provider and download matched shows
func (a *app) PullShows(ctx context.Context, p providers.Provider, pc *mpb.Progress) {
	if a.Config.Debug {
		log.Printf("Starting %s PullShows", p.Name())
	}
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		if a.Config.Debug {
			log.Printf("Starting %s PullShows", p.Name())
		}
	}()

	//log.Printf("Get shows list for %s", p.Name())
	pName := p.Name()
	shows, err := p.Shows(a.Config.WatchList)
	if err != nil {
		log.Printf("[%s] Can't get shows list of provider: %v", pName, err)
		return
	}

	seen := map[string]bool{}
	var bar *mpb.Bar
	if !a.Config.Headless {
		bar = pc.AddBar(int64(len(shows)),
			mpb.PrependDecorators(
				// simple name decorator
				decor.Name(left(p.Name(), 20), decor.WC{W: 20 + 1, C: decor.DidentRight}),
				decor.CountersNoUnit(" %3d/%3d", decor.WC{W: 5 + 1, C: decor.DidentRight}),
			))
	} else {
		log.Printf("[%s] %d shows available on server.", p.Name(), len(shows))
	}

	wg := sync.WaitGroup{}
showLoop:
	for id, s := range shows {
		if a.Config.Debug {
			log.Printf("PullShows %s, handling %d", p.Name(), id)
		}
		if _, ok := seen[s.ID]; ok {
			if !a.Config.Headless {
				bar.Increment()
			}
			continue
		}
		seen[s.ID] = true

		d := a.Config.Destinations[s.Destination]
		select {
		case <-ctx.Done():
			break showLoop
		default:
			if ctx.Err() != nil {
				break showLoop
			}
			if a.Config.Force || a.MustDownload(p, s, d) {
				if a.Config.Debug {
					log.Printf("PullShows %s, submitting %d", p.Name(), id)
				}
				wg.Add(1)
				a.SubmitDownload(ctx, &wg, p, s, d, pc, bar)
			} else {
				if !a.Config.Headless {
					bar.Increment()
				} else {
					log.Printf("[%s] %s already downloaded.", p.Name(), p.GetShowFileName(s))
				}
			}
		}
	}
	if !a.Config.Headless {
		bar.SetTotal(int64(len(shows)), true)
	}
	if ctx.Err() == nil {
		if a.Config.Debug {
			log.Println("PullShows waiting wg")
		}
		wg.Wait()
	}
	if a.Config.Debug {
		log.Println("Exit PullShows")
	}
}

// MustDownload check if the show isn't yet downloaded.
func (a *app) MustDownload(p providers.Provider, s *providers.Show, d string) bool {

	fn := filepath.Join(d, p.GetShowFileName(s))
	if _, err := os.Stat(fn); err == nil {
		return false
	}
	showPath := filepath.Join(d, p.GetShowFileNameMatcher(s))
	files, err := filepath.Glob(showPath)
	if err != nil {
		log.Fatalf("Can't glob %s: %v", showPath, err)
	}
	return len(files) == 0
}

func (a *app) SubmitDownload(ctx context.Context, wg *sync.WaitGroup, p providers.Provider, s *providers.Show, d string, pc *mpb.Progress, bar *mpb.Bar) {
	a.worker.Submit(func() {
		a.DownloadShow(ctx, wg, p, s, d, pc, bar)
	})
}

func (a *app) progressBar(ctx context.Context, fileBar *mpb.Bar, fn string, done chan bool) {
	// start := time.Now()
	if fileBar == nil {
		log.Fatal("FileBar should not been nil")
	}

	lastSize := int64(0)
	_ = lastSize
	t := time.NewTicker(500 * time.Millisecond)
	f := func() {
		s, err := os.Stat(fn)
		if err != nil {
			return
		}
		l := s.Size()
		fileBar.IncrInt64(l - lastSize)
		lastSize = l

		// if a.Config.Debug {
		// 	log.Printf("Bar tick %p, l:%d", fileBar, l)
		// }
	}
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case <-t.C:
			if ctx.Err() != nil {
				break loop
			}
			f()
		case <-done:
			t.Stop()
			f()
			break loop
		}
	}
	fileBar.SetTotal(100, true)
	if a.Config.Debug {
		log.Printf("Bar terminated %p", fileBar)
	}
}

var dlID = int64(0)

func (a *app) DownloadShow(ctx context.Context, wg *sync.WaitGroup, p providers.Provider, s *providers.Show, d string, pc *mpb.Progress, bar *mpb.Bar) {
	id := atomic.AddInt64(&dlID, 1)
	ctx, cancel := context.WithCancel(ctx)
	if a.Config.Headless || a.Config.Debug {
		log.Printf("Starting  DownloadShow %d", id)
	}

	// Make a context for DownloadShow
	files := []string{}
	shouldDeleteFile := false

	done := make(chan bool)

	var fileBar *mpb.Bar
	if !a.Config.Headless {
		bar.Increment()
		fileBar = pc.AddBar(100*1024*1024*1024,
			mpb.BarWidth(3),
			mpb.PrependDecorators(
				decor.Spinner([]string{"●∙∙", "∙●∙", "∙∙●", "∙●∙"}, decor.WCSyncSpace),
			),
			mpb.AppendDecorators(
				decor.AverageSpeed(decor.UnitKB, " %.1f", decor.WC{W: 15, C: decor.DidentRight}),
				decor.Name(p.GetShowFileName(s)),
			),
			mpb.BarRemoveOnComplete(),
		)
		if a.Config.Debug {
			log.Printf("Bar created %p", fileBar)
		}
	}

	fn := filepath.Join(d, p.GetShowFileName(s))
	defer func() {
		close(done)
		if shouldDeleteFile {
			for _, f := range files {
				log.Printf("Deleting incomplete files '%s'", f)
				os.Remove(f)
			}
		}
		cancel()
		if shouldDeleteFile && a.Config.Debug {
			log.Printf("[%s] %s terminated", p.Name(), fn)
		}
		if a.Config.Debug {
			log.Printf("DownloadShow %d terminated", id)
		}
		wg.Done()
	}()

	err := os.MkdirAll(filepath.Dir(fn), 0777)
	if err != nil {
		log.Println(err)
		return
	}

	url, err := p.GetShowStreamURL(s)
	if err != nil {
		log.Println(err)
		return
	}
	if strings.ToLower(filepath.Ext(url)) == ".m38u" {
		master, err := m3u8.NewMaster(url, a.getter)
		if err != nil {
			log.Println(err)
			return
		}
		url = master.BestQuality()
	}

	//log.Println("Download url: ", url)

	params := []string{
		"-loglevel", "quiet",
		"-hide_banner",
		"-i", url,
		"-metadata", "title=" + s.Title,
		"-metadata", "comment=" + s.Pitch,
		"-metadata", "show=" + s.Show,
		"-metadata", "channel=" + s.Channel,
		"-y",              // Override output file
		"-vcodec", "copy", // copy video
		"-acodec", "copy", // copy audio
		"-bsf:a", "aac_adtstoasc", // I don't know
		fn, // output file
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", params...)
	files = append(files, fn)

	if a.Config.Headless {
		log.Printf("[%s] Downloading %q", p.Name(), p.GetShowFileName(s))
	}

	if !a.Config.Headless {
		go a.progressBar(ctx, fileBar, fn, done)
	}

	err = cmd.Run()
	if err != nil {
		log.Println(err)
		shouldDeleteFile = true
		return
	}
	if ctx.Err() != nil {
		log.Println(ctx.Err())
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

	tbnStream, err := a.getter.Get(s.ThumbnailURL)
	if err != nil {
		log.Printf("[%s] Can't download %q's thumbnail: %v", p.Name(), p.GetShowFileName(s), err)

	}

	ws := []io.Writer{}
	tbnFile, err := os.Create(tbnFileName)
	if err != nil {
		log.Printf("[%s] Can't create %q's thumbnail: %v", p.Name(), p.GetShowFileName(s), err)
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
		log.Printf("[%s] Can't write %q's thumbnail: %v", p.Name(), p.GetShowFileName(s), err)
	}
	if a.Config.Headless || a.Config.Debug {
		log.Printf("[%s] %s downloaded.", p.Name(), fn)
	}
	return
}

type app struct {
	Config *Config
	Stop   chan bool
	ffmpeg string
	pb     *mpb.Progress // Progress bars
	worker *workers.WorkerPool
	getter getter
}

type getter interface {
	Get(uri string) (io.Reader, error)
}
