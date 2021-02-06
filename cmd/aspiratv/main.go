package main

import (
	"context"
	"path"
	"regexp"

	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	flag "github.com/spf13/pflag"

	"github.com/simulot/aspiratv/metadata/nfo"
	"github.com/simulot/aspiratv/mylog"
	"github.com/simulot/aspiratv/net/myhttp"
	"github.com/simulot/aspiratv/providers"
	"github.com/simulot/aspiratv/workers"
	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"

	_ "github.com/simulot/aspiratv/providers/artetv"
	_ "github.com/simulot/aspiratv/providers/francetv"
	_ "github.com/simulot/aspiratv/providers/gulli"
	"github.com/simulot/aspiratv/providers/matcher"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Config holds settings from configuration file
type config struct {
	Providers       map[string]ProviderConfig // Registered providers
	Force           bool                      // True to force reload medias
	Destinations    map[string]string         // Mapping of destination path
	ConfigFile      string                    // Name of configuration file
	WatchList       []*matcher.MatchRequest   // Slice of show matchers
	Headless        bool                      // When true, no progression bar
	ConcurrentTasks int                       // Number of concurrent downloads
	Provider        string                    // Provider for dowload command
	Destination     string                    // Destination folder for dowload command
	ShowPath        string                    // Imposed show's path
	LogFile         string                    // Log file
	WriteNFO        bool                      // True when NFO files to be written
	MaxAgedDays     int                       // Retrieve media younger than MaxAgedDays when non zero
	RetentionDays   int                       // Delete media from  series older than retention days.
	KeepBonus       bool                      // True to keep bonus
	LogLevel        string                    // ERROR,WARN,INFO,TRACE,DEBUG
	TitleFilter     string                    // ShowTitle or Episode title must match this regexp to be downloaded
	TitleExclude    string                    // ShowTitle and Episode title must not match this regexp to be downloaded
}

type app struct {
	Config config
	Stop   chan bool
	ffmpeg string
	pb     *mpb.Progress // Progress bars
	worker *workers.WorkerPool
	getter getter
	logger *mylog.MyLog
}

type getter interface {
	Get(ctx context.Context, uri string) (io.ReadCloser, error)
}

type logger interface {
	Printf(string, ...interface{})
}

func main() {
	var err error

	fmt.Printf("%s: %v, commit %v, built at %v\n", filepath.Base(os.Args[0]), version, commit, date)
	a := &app{
		Stop: make(chan bool),
	}

	// trap Ctrl+C and call cancel on the context
	ctx, cancel := context.WithCancel(context.Background())
	breakChannel := make(chan os.Signal, 1)
	signal.Notify(breakChannel, os.Interrupt)

	flag.StringVarP(&a.Config.LogLevel, "log-level", "l", "ERROR", "Log level (INFO,TRACE,ERROR,DEBUG)")
	flag.BoolVar(&a.Config.Force, "force", false, "Force media download.")
	flag.BoolVar(&a.Config.Headless, "headless", false, "Headless mode. Progression bars are not displayed.")
	flag.StringVar(&a.Config.ConfigFile, "config", "config.json", "Configuration file name.")
	flag.IntVarP(&a.Config.ConcurrentTasks, "max-tasks", "m", runtime.NumCPU(), "Maximum concurrent downloads at a time.")
	flag.StringVarP(&a.Config.Provider, "provider", "p", "", "Provider to be used with download command. Possible values : artetv, francetv, gulli")
	flag.StringVarP(&a.Config.Destination, "destination", "d", "", "Destination path for all shows.")
	flag.StringVarP(&a.Config.ShowPath, "show-path", "s", "", "Force show's path.")
	flag.StringVar(&a.Config.LogFile, "log", "", "Give the log file name.")
	// flag.IntVar(&a.Config.RetentionDays, "retention", 0, "Delete media older than retention days for the downloaded show.")
	flag.BoolVarP(&a.Config.WriteNFO, "write-nfo", "n", true, "Write NFO file for KODI,Emby,Plex...")
	flag.BoolVarP(&a.Config.KeepBonus, "keep-bonuses", "b", false, "Download bonuses when true")
	flag.IntVarP(&a.Config.MaxAgedDays, "max-aged", "a", 0, "Retrieve media younger than MaxAgedDays.")
	flag.StringVarP(&a.Config.TitleFilter, "title-filter", "f", "", "Showtitle or Episode title must satisfy regexp filter")
	flag.StringVarP(&a.Config.TitleExclude, "title-exclude", "e", "", "Showtitle and Episode title must not satisfy regexp filter")
	flag.Parse()

	consoleLogger := log.New(os.Stderr, "", log.LstdFlags)
	fileLogger := logger(nil)

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
		fileLogger = log.New(logFile, "", log.LstdFlags)
	}

	mylogger, err := mylog.NewLog(a.Config.LogLevel, consoleLogger, fileLogger)

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	a.logger = mylogger
	a.logger.Info().Printf("Command line paramters: %v", os.Args[1:])
	a.logger.Debug().Printf("Process PID: %d", os.Getpid())

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

	cmd := ""
	if len(flag.Args()) > 0 {
		cmd = flag.Arg(0)
	}

	a.Initialize(cmd)

	switch cmd {
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
			a.logger.Fatal().Printf("Can't sanitize path %q", k)
			os.Exit(1)
		}
		a.logger.Trace().Printf("Destination %q is expanded into %q", k, v)
		err = os.MkdirAll(v, 0755)
		if err != nil {
			a.logger.Fatal().Printf("Can't create destination directory %q: %s", v, err)
			os.Exit(1)
		}

		a.Config.Destinations[k] = v
	}
	if len(a.Config.ShowPath) > 0 {
		var err error
		a.Config.ShowPath, err = sanitizePath(a.Config.ShowPath)
		if err != nil {
			a.logger.Fatal().Printf("Can't sanitize path %q", a.Config.ShowPath)
			os.Exit(1)
		}
		a.logger.Trace().Printf("ShowPath set to  %q", a.Config.ShowPath)
	}
}

func sanitizePath(p string) (string, error) {
	return filepath.Abs(p)
}

// Download command
func (a *app) Download(ctx context.Context) {

	if len(flag.Args()) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	if len(a.Config.Destination) == 0 && len(a.Config.ShowPath) == 0 {
		a.logger.Fatal().Printf("Either --destination or --show-path parameter is mandatory for download operation")
	}

	// if (len(a.Config.Destinations) > 0) != (len(a.Config.ShowPath) > 0) {
	// 	a.logger.Fatal().Printf("Choose one of paramters --destination or --show-path for download operation")
	// }

	if a.Config.TitleFilter != "" {
		if _, err := regexp.Compile(a.Config.TitleFilter); err != nil {
			a.logger.Fatal().Printf("Can't use the title-filter: %q", err)
		}
	}

	if a.Config.TitleExclude != "" {
		if _, err := regexp.Compile(a.Config.TitleExclude); err != nil {
			a.logger.Fatal().Printf("Can't use the title-exclude: %q", err)
		}
	}

	a.Config.Destinations = map[string]string{
		"DL": os.ExpandEnv(a.Config.Destination),
	}

	a.Config.ShowPath = os.ExpandEnv(a.Config.ShowPath)

	a.CheckPaths()
	a.Config.WatchList = []*matcher.MatchRequest{}

	var filter, exclude matcher.Filter

	if a.Config.TitleFilter != "" {
		filter = matcher.Filter{Regexp: regexp.MustCompile(a.Config.TitleFilter)}
	}
	if a.Config.TitleExclude != "" {
		exclude = matcher.Filter{Regexp: regexp.MustCompile(a.Config.TitleExclude)}
	}

	for dl := 1; dl < flag.NArg(); dl++ {
		mr := matcher.MatchRequest{
			Destination:   "DL",
			Show:          strings.ToLower(flag.Arg(dl)),
			Provider:      a.Config.Provider,
			MaxAgedDays:   a.Config.MaxAgedDays,
			RetentionDays: a.Config.RetentionDays,
			TitleFilter:   filter,
			TitleExclude:  exclude,
			ShowRootPath:  a.Config.ShowPath,
			KeepBonus:     a.Config.KeepBonus,
		}
		a.Config.WatchList = append(a.Config.WatchList, &mr)

	}
	a.worker = workers.New(ctx, a.Config.ConcurrentTasks, a.logger) //TODO
	a.getter = myhttp.DefaultClient

	if a.Config.Provider == "" {
		a.logger.Fatal().Printf("Missing --provider PROVIDERNAME flag")
		os.Exit(1)
	}

	p, ok := providers.List()[a.Config.Provider]
	if !ok {
		a.logger.Fatal().Printf("Unknown provider %q", a.Config.Provider)
		os.Exit(1)
	}
	p.Configure(providers.Config{
		Log: a.logger,
	})

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
	a.worker = workers.New(ctx, a.Config.ConcurrentTasks, a.logger) // TODO
	a.getter = myhttp.DefaultClient

	pc := a.getProgres(ctx)

	activeProviders := int64(0)
	for _, p := range providers.List() {
		if a.Config.IsProviderActive(p.Name()) {
			activeProviders++
			p.Configure(providers.Config{
				Log: a.logger,
			})
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
					a.logger.Trace().Printf("[%s] Pulling shows", p.Name())
					a.PullShows(ctx, p, pc)
					wg.Done()
					a.logger.Trace().Printf("[%s] Pulling completed", p.Name())
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
	a.logger.Debug().Printf("End of providerLoop")
	a.worker.Stop()
	a.logger.Debug().Printf("Workers stop confirmed")
	a.logger.Debug().Printf("End of Run")
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
	a.logger.Debug().Printf("[%s] Starting PullShows", p.Name())
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

	a.logger.Trace().Printf("[%s] Get shows list", p.Name())
	seen := map[string]bool{}
	wg := sync.WaitGroup{}

	showCount := int64(0)
showLoop:

	for m := range p.MediaList(ctx, a.Config.WatchList) {
		a.logger.Trace().Printf("[%s] Get id  %s", p.Name(), m.ID)

		if _, ok := seen[m.ID]; ok {
			continue
		}
		seen[m.ID] = true

		if m.Match.ShowRootPath == "" {
			m.ShowPath = path.Join(a.Config.Destinations[m.Match.Destination], nfo.FileNameCleaner(m.Metadata.GetMediaInfo().Showtitle))
		} else {
			m.ShowPath = m.Match.ShowRootPath
		}

		mediaBaseName := filepath.Base(m.Metadata.GetMediaPath(m.ShowPath))

		select {
		case <-ctx.Done():
			a.logger.Trace().Printf("[%s] Context done, received %s", p.Name(), ctx.Err())
			break showLoop
		default:
			if !m.Metadata.Accepted(m.Match) {
				a.logger.Trace().Printf("[%s] %s is filtered out.", p.Name(), mediaBaseName)
			} else if a.Config.Force || a.MustDownload(ctx, p, m) {
				a.logger.Trace().Printf("[%s] Download of %q submitted", p.Name(), mediaBaseName)
				showCount++
				if !a.Config.Headless {
					providerBar.SetTotal(showCount, false)
				}
				a.SubmitDownload(ctx, &wg, p, m, pc, providerBar)
			} else {
				a.logger.Trace().Printf("[%s] %s already downloaded.", p.Name(), mediaBaseName)
			}
			if ctx.Err() != nil {
				a.logger.Debug().Printf("[%s] PullShows received %s", p.Name(), ctx.Err())
				break showLoop
			}
		}
	}
	if !a.Config.Headless {
		providerBar.SetTotal(showCount, showCount == 0)
	}
	a.logger.Debug().Printf("Waiting end of PullShows loop")

	// Wait for submitted jobs to be terminated
	wg.Wait()

	if !a.Config.Headless {
		providerBar.SetTotal(showCount, true)
	}
	a.logger.Debug().Printf("Exit PullShows")
}

// MustDownload check if the show isn't yet downloaded.
func (a *app) MustDownload(ctx context.Context, p providers.Provider, m *providers.Media) bool {
	mediaPath := m.Metadata.GetMediaPath(m.ShowPath)
	mediaExists, err := fileExists(mediaPath)
	if mediaExists {
		return false
	}

	mediaPath = m.Metadata.GetMediaPathMatcher(m.ShowPath)
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

func (a *app) SubmitDownload(ctx context.Context, wg *sync.WaitGroup, p providers.Provider, m *providers.Media, pc *mpb.Progress, bar *mpb.Bar) {
	wg.Add(1)
	go a.worker.Submit(func() {
		a.DownloadShow(ctx, p, m, pc)
		if bar != nil {
			bar.Increment()
		}
	}, wg)
}
