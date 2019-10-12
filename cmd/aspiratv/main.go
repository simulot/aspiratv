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

	"github.com/simulot/aspiratv/net/myhttp"
	"github.com/simulot/aspiratv/providers"
	"github.com/simulot/aspiratv/workers"
	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"

	_ "github.com/simulot/aspiratv/providers/artetv"
	_ "github.com/simulot/aspiratv/providers/francetv"
	// _ "github.com/simulot/aspiratv/providers/gulli"
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
	MaxAgedDays     int    // Retrieve media younger than MaxAgedDays when non zero
	RetentionDays   int    // Delete media from  series older than retention days.
	LogFile         string // Log file
	WriteNFO        bool   // True when NFO files to be written
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
	flag.IntVar(&a.Config.MaxAgedDays, "max-aged", 0, "Retrieve media younger than MaxAgedDays.")
	flag.IntVar(&a.Config.RetentionDays, "retention", 0, "Delete media older than retention days for the downloaded show.")
	flag.BoolVar(&a.Config.WriteNFO, "write-nfo", true, "Write NFO file for KODI,Emby,Plex...")
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
			Destination:   "DL",
			Show:          strings.ToLower(flag.Arg(1)),
			Provider:      a.Config.Provider,
			MaxAgedDays:   a.Config.MaxAgedDays,
			RetentionDays: a.Config.RetentionDays,
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

	for m := range p.MediaList(ctx, a.Config.WatchList) {
		if _, ok := seen[m.ID]; ok {
			continue
		}
		seen[m.ID] = true

		select {
		case <-ctx.Done():
			break showLoop
		default:

			if a.Config.Force || a.MustDownload(ctx, p, m) {
				if a.Config.Headless {
					log.Printf("[%s] Download of %q submitted", p.Name(), filepath.Base(m.Metadata.GetMediaPath(a.Config.Destinations[m.Match.Destination])))
				}
				showCount++
				if !a.Config.Headless {
					providerBar.SetTotal(showCount, false)
				}
				a.SubmitDownload(ctx, &wg, p, m, pc, providerBar)
			} else {
				if a.Config.Headless {
					log.Printf("[%s] %s already downloaded.", p.Name(), filepath.Base(m.Metadata.GetMediaPath(a.Config.Destinations[m.Match.Destination])))
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
func (a *app) MustDownload(ctx context.Context, p providers.Provider, m *providers.Media) bool {
	mediaPath := m.Metadata.GetMediaPath(a.Config.Destinations[m.Match.Destination])
	mediaExists, err := fileExists(mediaPath)
	if mediaExists {
		return false
	}

	mediaPath = m.Metadata.GetMediaPathMatcher(a.Config.Destinations[m.Match.Destination])
	files, err := filepath.Glob(mediaPath)
	if err != nil {
		log.Fatalf("Can't glob %s: %v", mediaPath, err)
	}
	return len(files) == 0
}

func fileExists(p string) (bool, error) {
	_, err := os.Stat(p)
	if os.IsExist(err) {
		return false, err
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, nil
}
