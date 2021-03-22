package main

import (
	"context"

	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"

	flag "github.com/spf13/pflag"

	"github.com/simulot/aspiratv/mylog"
	"github.com/simulot/aspiratv/providers"

	"github.com/simulot/aspiratv/matcher"
	_ "github.com/simulot/aspiratv/providers/artetv"
	_ "github.com/simulot/aspiratv/providers/francetv"
	_ "github.com/simulot/aspiratv/providers/gulli"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type app struct {
	// CLI flags
	Settings        providers.Settings   // Application global settings
	Matcher         matcher.MatchRequest // Matcher for the command line
	ConfigFile      string               // Name of configuration file
	Headless        bool                 // When true, no progression bar
	ConcurrentTasks int                  // Number of concurrent downloads
	LogLevel        string               // ERROR,WARN,INFO,TRACE,DEBUG
	LogFile         string               // Log file
	WaitDebugger    bool                 // When true, the PID is displayed, and wait for ENTER key

	// State
	Stop   chan bool
	ffmpeg string
	// pb     *mpb.Progress // Progress bars
	// worker     *workers.WorkerPool
	// getter     getter
	logger     *mylog.MyLog
	fsRun      *flag.FlagSet
	fsDownload *flag.FlagSet

	// Progression bars
	BarContainer *barContainer
}

// type getter interface {
// 	Get(ctx context.Context, uri string) (io.ReadCloser, error)
// }

type logger interface {
	Printf(string, ...interface{})
}

func main() {
	var err error
	var command string

	fmt.Printf("%s: %v, commit %v, built at %v\n", filepath.Base(os.Args[0]), version, commit, date)
	a := &app{
		Stop: make(chan bool),
	}

	// trap Ctrl+C and call cancel on the context
	ctx, cancel := context.WithCancel(context.Background())
	breakChannel := make(chan os.Signal, 1)
	signal.Notify(breakChannel, os.Interrupt)

	a.SetFlags()

	switch {
	case len(os.Args) > 1 && os.Args[1] == "download":
		command = "download"
		err = a.fsDownload.Parse(os.Args[2:])
	case len(os.Args) > 1 && os.Args[1] == "run":
		command = "run"
		err = a.fsRun.Parse(os.Args[2:])
	case len(os.Args) > 1 && os.Args[1] == "help":
		command = "help"
		err = a.fsRun.Parse(os.Args[2:])
	default:
		command = "run"
		err = a.fsRun.Parse(os.Args[1:])
	}

	if err != nil {
		fmt.Println(err)
		a.Usage()
	}

	if a.WaitDebugger {
		fmt.Println("PID:", os.Getpid())
		fmt.Print("Press Enter to continue ")
		fmt.Scanln()
	}

	logFile := os.Stdout
	if len(a.LogFile) > 0 {
		var err error
		logFile, err = os.Create(a.LogFile)
		if err != nil {
			log.Printf("Can't create log file: %q", err)
			os.Exit(1)
		}
		defer func() {
			logFile.Sync()
			logFile.Close()
		}()
	}

	mylogger, err := mylog.NewLog(a.LogLevel, log.New(logFile, "", log.LstdFlags))

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
			a.logger.Info().Printf("^C pressed...")
			cancel()
		case <-ctx.Done():
			return
		}
	}()

	a.Initialize(command)

	switch command {
	case "download":
		if len(a.fsDownload.Args()) > 0 {
			a.Download(ctx, a.fsDownload.Args()[0])
		} else {
			a.Exit("Missing show name")
		}
	case "run":
		a.Run(ctx)
	case "help":
		a.Usage()
	}

	a.logger.Info().Printf("Program stopped")
}

func (a *app) Exit(message string) {
	fmt.Println()
	fmt.Println(message)
	fmt.Println()
	a.Usage()
	os.Exit(1)
}
