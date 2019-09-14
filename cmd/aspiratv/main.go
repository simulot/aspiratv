package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
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

type app struct {
	Config Config
	Stop   chan bool
	ffmpeg string
	pb     *mpb.Progress // Progress bars
	worker *workers.WorkerPool
	getter getter
}

type getter interface {
	Get(ctx context.Context, uri string) (io.ReadCloser, error)
}

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

	flag.BoolVar(&a.Config.Debug, "debug", false, "Debug mode.")
	flag.BoolVar(&a.Config.Force, "force", false, "Force media download.")
	flag.BoolVar(&a.Config.Headless, "headless", false, "Headless mode. Progression bars are not displayed.")
	flag.StringVar(&a.Config.ConfigFile, "config", "config.json", "Configuration file name.")
	flag.IntVar(&a.Config.ConcurrentTasks, "max-tasks", runtime.NumCPU(), "Maximum concurrent downloads at a time.")
	flag.StringVar(&a.Config.Provider, "provider", "", "Provider to be used with download command. Possible values : artetv,francetv,gulli")
	flag.StringVar(&a.Config.Destination, "destination", "", "Provider to be used with download command. Possible values : artetv,francetv,gulli")
	flag.StringVar(&a.Config.LogFile, "log", "", "Give the log file name. When empty, no log.")
	flag.Parse()

	if len(a.Config.LogFile) > 0 {
		logFile, err := os.Create(a.Config.LogFile)
		if err != nil {
			log.Printf("Can't create log file: %q", err)
			os.Exit(1)
		}
		defer func() {
			logFile.Sync()
			logFile.Close()
		}()
		log.SetOutput(logFile)
	} else {
		if a.Config.Headless {
			log.SetOutput(os.Stdout)
		} else {
			log.SetOutput(ioutil.Discard)
		}
	}

	a.Initialize()
	if len(os.Args) < 1 {
		flag.Usage()
		os.Exit(1)
	}
	switch flag.Arg(0) {
	case "download":
		a.Download(ctx)
	default:
		a.Run(ctx)
	}
}

func (a *app) Download(ctx context.Context) {
	if len(flag.Args()) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	a.Config.Destinations = map[string]string{
		"DL": os.ExpandEnv(a.Config.Destination),
	}

	a.Config.WatchList = []*providers.MatchRequest{
		&providers.MatchRequest{
			Destination: "DL",
			Show:        strings.ToLower(flag.Arg(1)),
			Provider:    a.Config.Provider,
		},
	}
	a.worker = workers.New(ctx, a.Config.ConcurrentTasks, a.Config.Debug)
	a.getter = http.DefaultClient

	if a.Config.Provider == "" {
		log.Println("Missing -provider PROVIDERNAME flag")
		os.Exit(1)
	}

	p, ok := providers.List()[a.Config.Provider]
	if !ok {
		log.Printf("Unknown provider %q", a.Config.Provider)
		os.Exit(1)
	}

	pc := a.getProgres(ctx)

	a.PullShows(ctx, p, pc)
	if !a.Config.Headless {
		pc.Wait()
	}
}

func (a *app) getProgres(ctx context.Context) *mpb.Progress {
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
	return pc
}

func (a *app) Run(ctx context.Context) {

	a.worker = workers.New(ctx, a.Config.ConcurrentTasks, a.Config.Debug)
	a.getter = http.DefaultClient

	pc := a.getProgres(ctx)

	activeProviders := int64(0)
	for _, p := range providers.List() {
		if a.Config.IsProviderActive(p.Name()) {
			activeProviders++
		}
	}

	wg := sync.WaitGroup{}
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
				wg.Add(1)
				go func(p providers.Provider) {
					if a.Config.Headless {
						log.Printf("[%s] Pulling shows", p.Name())
					}
					a.PullShows(ctx, p, pc)
					wg.Done()
					log.Printf("[%s] Pulling completed", p.Name())
				}(p)
			}
		}
	}

	if ctx.Err() == nil {
		wg.Wait()
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

var nbPuller = int32(0)

// PullShows pull provider and download matched shows
func (a *app) PullShows(ctx context.Context, p providers.Provider, pc *mpb.Progress) {
	if a.Config.Debug {
		log.Printf("Starting %s PullShows", p.Name())
		p.DebugMode(true)
	}
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		if a.Config.Debug {
			log.Printf("Starting %s PullShows", p.Name())
		}
	}()

	var providerBar *mpb.Bar

	if !a.Config.Headless {
		providerBar = pc.AddBar(0,
			mpb.BarWidth(50),
			mpb.PrependDecorators(
				decor.Name("Pulling "+p.Name(), decor.WC{W: 20, C: decor.DidentRight}),
			),
			mpb.AppendDecorators(
				decor.OnComplete(decor.Counters(0, "  %d/%d"), "completed"),
			),
		)
		providerBar.SetPriority(int(atomic.AddInt32(&nbPuller, 1)))
		if a.Config.Debug {
			log.Printf("Provider Bar created %p", providerBar)
		}
	}

	//log.Printf("Get shows list for %s", p.Name())
	seen := map[string]bool{}
	wg := sync.WaitGroup{}

	showCount := int64(0)
showLoop:
	for s := range p.Shows(ctx, a.Config.WatchList) {
		if _, ok := seen[s.ID]; ok {
			continue
		}
		seen[s.ID] = true

		select {
		case <-ctx.Done():
			break showLoop
		default:
			d := a.Config.Destinations[s.Destination]
			if a.Config.Force || a.MustDownload(ctx, p, s, d) {
				if a.Config.Debug {
					log.Printf("[%s] submitting %d", p.Name(), showCount)
				}
				showCount++
				if !a.Config.Headless {
					providerBar.SetTotal(showCount+1, false)
				}

				wg.Add(1)
				a.SubmitDownload(ctx, &wg, p, s, d, pc, providerBar)
			} else {
				if a.Config.Headless {
					log.Printf("[%s] %s already downloaded.", p.Name(), p.GetShowFileName(ctx, s))
				}
			}
			if ctx.Err() != nil {
				break showLoop
			}

		}
	}
	if a.Config.Debug {
		log.Println("Waiting end of PullShows loop")
	}
	shutDownChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(shutDownChan)
	}()

	// gracefull wait the end of show download or cancellation
	select {
	case <-ctx.Done():
	case <-shutDownChan:
	}

	if !a.Config.Headless {
		providerBar.SetTotal(showCount, true)
	}
	if a.Config.Debug {
		log.Println("Exit PullShows")
	}
}

// MustDownload check if the show isn't yet downloaded.
func (a *app) MustDownload(ctx context.Context, p providers.Provider, s *providers.Show, d string) bool {

	fn := filepath.Join(d, p.GetShowFileName(ctx, s))
	if _, err := os.Stat(fn); err == nil {
		return false
	}
	showPath := filepath.Join(d, p.GetShowFileNameMatcher(ctx, s))
	files, err := filepath.Glob(showPath)
	if err != nil {
		log.Fatalf("Can't glob %s: %v", showPath, err)
	}
	return len(files) == 0
}

func (a *app) SubmitDownload(ctx context.Context, wg *sync.WaitGroup, p providers.Provider, s *providers.Show, d string, pc *mpb.Progress, bar *mpb.Bar) {
	go a.worker.Submit(func() {
		a.DownloadShow(ctx, p, s, d, pc)
		if !a.Config.Headless {
			bar.Increment()
		}
		wg.Done()
	})
}

func (a *app) progressBar(ctx context.Context, fileBar *mpb.Bar, fn string, done chan bool) {
	// start := time.Now()
	if fileBar == nil {
		log.Fatal("FileBar should not been nil")
	}

	lastSize := int64(0)
	t := time.NewTicker(500 * time.Millisecond)
	f := func() {
		s, err := os.Stat(fn)
		if err != nil {
			return
		}
		l := s.Size()
		fileBar.IncrInt64(l - lastSize)
		lastSize = l
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

var dlID = int32(0)

func (a *app) DownloadShow(ctx context.Context, p providers.Provider, s *providers.Show, d string, pc *mpb.Progress) {
	id := atomic.AddInt32(&dlID, 1)
	ctx, cancel := context.WithCancel(ctx)
	if a.Config.Debug {
		log.Printf("Starting  DownloadShow %d", id)
	}

	// Make a context for DownloadShow
	files := []string{}
	shouldDeleteFile := false

	done := make(chan bool)

	var fileBar *mpb.Bar
	if !a.Config.Headless {
		fileBar = pc.AddBar(100*1024*1024*1024,
			mpb.BarWidth(3),
			mpb.PrependDecorators(
				decor.Spinner([]string{"●∙∙", "∙●∙", "∙∙●", "∙●∙"}, decor.WCSyncSpace),
			),
			mpb.AppendDecorators(
				decor.AverageSpeed(decor.UnitKB, " %.1f", decor.WC{W: 15, C: decor.DidentRight}),
				decor.Name(filepath.Base(p.GetShowFileName(ctx, s))),
			),
			mpb.BarRemoveOnComplete(),
		)
		if a.Config.Debug {
			log.Printf("Bar created %p", fileBar)
		}
		fileBar.SetPriority(int(100 + dlID))
	}

	fn := filepath.Join(d, p.GetShowFileName(ctx, s))
	defer func() {
		close(done)
		if shouldDeleteFile {
			for _, f := range files {
				log.Printf("[%s] %s is cancelled.", p.Name(), p.GetShowFileName(ctx, s))
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
		if !a.Config.Headless {
			fileBar.SetTotal(1, true)
		}
	}()

	if a.Config.Debug {
		log.Printf("Download stream to: %q", fn)
	}
	err := os.MkdirAll(filepath.Dir(fn), 0777)
	if err != nil {
		log.Println(err)
		return
	}

	url, err := p.GetShowStreamURL(ctx, s)
	if err != nil {
		log.Println(err)
		return
	}
	if len(url) == 0 {
		log.Printf("[%s] Can't get url from %s.", p.Name(), p.GetShowFileName(ctx, s))
		return
	}
	if strings.ToLower(filepath.Ext(url)) == ".m38u" {
		master, err := m3u8.NewMaster(ctx, url, a.getter)
		if err != nil {
			log.Println(err)
			return
		}
		url = master.BestQuality()
	}

	if a.Config.Debug {
		log.Println("Download url: ", url)
	}

	params := []string{
		"-loglevel", "error",
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

	if a.Config.Debug {
		log.Printf("[%s] Downloading %q", p.Name(), p.GetShowFileName(ctx, s))
	}

	if !a.Config.Headless {
		go a.progressBar(ctx, fileBar, fn, done)
	}

	err = cmd.Run()
	if err != nil {
		if err, ok := err.(*exec.ExitError); ok {
			log.Printf("[%s] FFMEPG exits with error:\n%s", p.Name(), err.Stderr)
		}
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
		log.Printf("[%s] Can't download %q's thumbnail: %v", p.Name(), p.GetShowFileName(ctx, s), err)
	}
	ws := []io.Writer{}
	tbnFile, err := os.Create(tbnFileName)
	if err != nil {
		log.Printf("[%s] Can't create %q's thumbnail: %v", p.Name(), p.GetShowFileName(ctx, s), err)
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
		log.Printf("[%s] Can't write %q's thumbnail: %v", p.Name(), p.GetShowFileName(ctx, s), err)
	}
	if a.Config.Headless || a.Config.Debug {
		log.Printf("[%s] %s downloaded.", p.Name(), p.GetShowFileName(ctx, s))
	}
	return
}
