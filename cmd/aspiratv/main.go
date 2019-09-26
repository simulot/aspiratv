package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/simulot/aspiratv/download"

	"github.com/simulot/aspiratv/net/myhttp"
	"github.com/simulot/aspiratv/providers"
	"github.com/simulot/aspiratv/workers"
	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"

	_ "github.com/simulot/aspiratv/providers/artetv"
	_ "github.com/simulot/aspiratv/providers/francetv"
	_ "github.com/simulot/aspiratv/providers/gulli"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Config holds settings from configuration file
type config struct {
	Debug           bool                      // Verbose Log output
	Force           bool                      // True to force reload medias
	Destinations    map[string]string         // Mapping of destination path
	ConfigFile      string                    // Name of configuration file
	WatchList       []*providers.MatchRequest // Slice of show matchers
	Headless        bool                      // When true, no progression bar
	ConcurrentTasks int                       // Number of concurrent downloads
	Providers       map[string]ProviderConfig
	Provider        string // Provider for dowload command
	Destination     string // Destination folder for dowload command
	LogFile         string // Log file

}

type app struct {
	Config config
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

	if a.Config.Debug {
		fmt.Print("PID: ", os.Getpid(), ", press enter to continue")
		var input string
		fmt.Scanln(&input)
	}

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

func (a *app) CheckPaths() {
	for k, v := range a.Config.Destinations {
		var err error
		v, err = sanitizePath(v)
		if err != nil {
			log.Printf("Destination %q is unsafe", k)
			os.Exit(1)
		}
		if a.Config.Debug {
			log.Printf("Destination %q is expanded into %q", k, v)
		}
		err = os.MkdirAll(v, 0755)
		if err != nil {
			log.Printf("Can't create destination directory %q: %s", v, err)
			os.Exit(1)
		}

		a.Config.Destinations[k] = v
	}
}

func sanitizePath(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if !filepath.IsAbs(p) {
		p = filepath.Clean(p)
	}
	if err != nil {
		return "", fmt.Errorf("Unsafe path: %w", err)
	}
	if !strings.Contains(abs, p) {
		return "", errors.New("Unsafe path")
	}
	return abs, nil
}

func (a *app) Download(ctx context.Context) {
	if len(flag.Args()) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	a.Config.Destinations = map[string]string{
		"DL": os.ExpandEnv(a.Config.Destination),
	}
	a.CheckPaths()
	a.Config.WatchList = []*providers.MatchRequest{
		&providers.MatchRequest{
			Destination: "DL",
			Show:        strings.ToLower(flag.Arg(1)),
			Provider:    a.Config.Provider,
		},
	}
	a.worker = workers.New(ctx, a.Config.ConcurrentTasks, a.Config.Debug)
	a.getter = myhttp.DefaultClient

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
	a.CheckPaths()
	a.worker = workers.New(ctx, a.Config.ConcurrentTasks, a.Config.Debug)
	a.getter = myhttp.DefaultClient

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
		log.Printf("[%s] Starting PullShows", p.Name())
		p.DebugMode(true)
	}
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
	}()

	var providerBar *mpb.Bar

	if !a.Config.Headless {
		providerBar = pc.AddBar(0,
			mpb.BarWidth(12),
			mpb.AppendDecorators(
				decor.OnComplete(decor.Counters(0, "  %3d/%3d"), "completed"),
				decor.Name("      Pulling "+p.Name()),
			),
		)
		providerBar.SetPriority(int(atomic.AddInt32(&nbPuller, 1)))
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
					providerBar.SetTotal(showCount, false)
				}
				a.SubmitDownload(ctx, &wg, p, s, d, pc, providerBar)
			} else {
				if a.Config.Headless {
					log.Printf("[%s] %s already downloaded.", p.Name(), providers.GetShowFileName(ctx, s))
				}
			}
			if ctx.Err() != nil {
				break showLoop
			}
		}
	}
	if !a.Config.Headless {
		providerBar.SetTotal(showCount, showCount == 0)
	}
	if a.Config.Debug {
		log.Println("Waiting end of PullShows loop")
	}

	// Wait for submitted jobs to be terminated
	wg.Wait()

	if !a.Config.Headless {
		providerBar.SetTotal(showCount, true)
	}
	if a.Config.Debug {
		log.Println("Exit PullShows")
	}
}

// MustDownload check if the show isn't yet downloaded.
func (a *app) MustDownload(ctx context.Context, p providers.Provider, s *providers.Show, d string) bool {

	fn := filepath.Join(d, providers.GetShowFileName(ctx, s))
	if _, err := os.Stat(fn); err == nil {
		return false
	}
	showPath := filepath.Join(d, providers.GetShowFileNameMatcher(ctx, s))
	files, err := filepath.Glob(showPath)
	if err != nil {
		log.Fatalf("Can't glob %s: %v", showPath, err)
	}
	return len(files) == 0
}

func (a *app) SubmitDownload(ctx context.Context, wg *sync.WaitGroup, p providers.Provider, s *providers.Show, d string, pc *mpb.Progress, bar *mpb.Bar) {
	wg.Add(1)
	go a.worker.Submit(func() {
		if ctx.Err() == nil {
			a.DownloadShow(ctx, p, s, d, pc)
		}
		wg.Done()
		if !a.Config.Headless {
			bar.Increment()
		}
	})
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
		// p.totalCount = count
		// fmt.Printf("%.1f%%\n", float64(count)/float64(p.totalCount)*100.0)
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

	done := make(chan bool)

	fn := filepath.Join(d, providers.GetShowFileName(ctx, s))
	if a.Config.Debug {
		log.Printf("[%s] Downloading into file: %q", p.Name(), fn)
	}
	defer func() {
		close(done)
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
		// if err, ok := err.(*exec.ExitError); ok {
		log.Printf("[%s] FFMEPG exits with error:\n%s", p.Name(), err)
		// }
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
