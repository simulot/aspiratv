package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	flag "github.com/spf13/pflag"
)

func (a *app) SetFlags() {
	a.SetRunFlags()
	a.SetDownloadFlags()
}

func (a *app) SetRunFlags() {
	a.fsRun = flag.NewFlagSet("run", flag.ExitOnError)
	a.fsRun.StringVar(&a.ConfigFile, "config", "config.json", "Configuration file name.")
	a.fsRun.Usage = func() {
		fmt.Println("Command run: downloawd new shows listed into configuration file in the watchlist")
		fmt.Println()
		fmt.Println(filepath.Base(os.Args[0]), " run [ options... ]")
		fmt.Println(filepath.Base(os.Args[0]), " [ options... ]")
		fmt.Println()
		fmt.Println("  example:  ", filepath.Base(os.Args[0]), "run", "--log aspiratv.log")
		fmt.Println()
		fmt.Println("  options:")
		a.fsRun.PrintDefaults()
		fmt.Println()
	}

	a.addCommonFlags(a.fsRun)
}

func (a *app) SetDownloadFlags() {
	a.fsDownload = flag.NewFlagSet("download", flag.ExitOnError)
	a.fsDownload.StringVarP(&a.Matcher.Provider, "provider", "p", "", "Provider to be used with download command. Possible values : artetv, francetv, gulli (mandatory).")
	a.fsDownload.StringVarP(&a.Matcher.ShowRootPath, "show-path", "s", "", "Show's path (mandatory).")
	a.fsDownload.BoolVar(&a.Matcher.Force, "force", false, "Force media download even when present on the show-path.")
	// a.fsDownload.StringVarP(&a.Matcher.Destination, "destination", "d", "", "Destination path for all shows.")
	a.fsDownload.IntVar(&a.Matcher.RetentionDays, "retention", 0, "Delete media older than retention days for the downloaded show.")
	a.fsDownload.BoolVarP(&a.Matcher.KeepBonus, "keep-bonuses", "b", false, "Download bonuses when true")
	a.fsDownload.IntVarP(&a.Matcher.MaxAgedDays, "max-aged", "a", 0, "Retrieve media younger than MaxAgedDays.")
	a.fsDownload.VarP(&a.Matcher.TitleFilter, "title-filter", "f", "Showtitle or Episode title must satisfy regexp filter")
	a.fsDownload.VarP(&a.Matcher.TitleExclude, "title-exclude", "e", "Showtitle and Episode title must not satisfy regexp filter")
	a.fsDownload.Var(&a.Matcher.ShowNameTemplate, "name-template", "Show name file template")
	a.fsDownload.Var(&a.Matcher.SeasonPathTemplate, "season-template", "Season directory template")

	a.fsDownload.Usage = func() {
		fmt.Println("Command download: download show with given options")
		fmt.Println()
		fmt.Println(filepath.Base(os.Args[0]), " download --provider PROVIDER --show-path PATH [ options... ] \"show name\"")
		fmt.Println()
		fmt.Println("  example:  ", filepath.Base(os.Args[0]), " download --provider francetv --show-path {$HOME}/Videos/Animes/  \"lapins crétins\"")
		fmt.Println()
		fmt.Println("  options:")
		a.fsDownload.PrintDefaults()
	}
	a.addCommonFlags(a.fsDownload)
}

func (a *app) addCommonFlags(fs *flag.FlagSet) {
	fs.StringVarP(&a.LogLevel, "log-level", "l", "ERROR", "Log level (INFO,TRACE,ERROR,DEBUG)")
	fs.BoolVar(&a.Headless, "headless", false, "Headless mode. Progression bars are not displayed.")
	fs.IntVarP(&a.ConcurrentTasks, "max-tasks", "m", runtime.NumCPU(), "Maximum concurrent downloads at a time.")
	fs.StringVar(&a.LogFile, "log", "", "Give the log file name.")
	fs.BoolVar(&a.WaitDebugger, "debugger", false, "Wait for debugger")
	fs.MarkHidden("debugger")
}

func (a *app) Usage() {
	fmt.Printf("Usage of %s:\n\n", os.Args[0])
	a.fsRun.Usage()
	a.fsDownload.Usage()
	os.Exit(1)
}
