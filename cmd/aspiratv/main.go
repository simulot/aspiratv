package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
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

	cliConfig := &Config{}

	flag.BoolVar(&cliConfig.Service, "service", false, "Run as service.")
	flag.BoolVar(&cliConfig.Debug, "debug", false, "Debug mode.")
	flag.BoolVar(&cliConfig.Force, "force", false, "Force media download.")
	flag.BoolVar(&cliConfig.Headless, "headless", false, "Headless mode. Progression bars are not displayed.")
	flag.StringVar(&cliConfig.ConfigFile, "config", "config.json", "Configuration file name.")
	flag.Parse()

	log.SetOutput(os.Stderr)

	a.Initialize(cliConfig)
	if a.Config.Service {
		a.RunAsService()
	} else {
		a.RunOnce()
	}
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
	a.worker = workers.New()
	a.getter = http.DefaultClient
}

func (a *app) RunOnce() {
	a.RunAll()
	a.worker.Stop()
	log.Println("Job(s) are done!")
}

func (a *app) RunAsService() {
	for {
		a.RunAll()
		s := time.Duration(a.Config.PullInterval) + time.Duration(rand.Intn(int(time.Duration(a.Config.PullInterval))/4))
		//log.Printf("Sleeping until %s\n", time.Now().Add(s).Format("15:04:05"))
		time.Sleep(s)
	}
}

func (a *app) RunAll() {
	pc := mpb.New(
		mpb.WithWidth(64),
		mpb.ContainerOptOnCond(
			mpb.WithOutput(nil),
			func() bool {
				return a.Config.Headless
			},
		))
	activeProviders := int64(0)
	for _, p := range providers.List() {
		if a.Config.IsProviderActive(p.Name()) {
			activeProviders++
		}
	}
	for _, p := range providers.List() {
		if a.Config.IsProviderActive(p.Name()) {
			a.PullShows(p, pc)
		}
	}
	pc.Wait()
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
func (a *app) PullShows(p providers.Provider, pc *mpb.Progress) {

	//log.Printf("Get shows list for %s", p.Name())
	pName := p.Name()
	shows, err := p.Shows(a.Config.WatchList)
	if err != nil {
		log.Printf("[%s] Can't get shows list of provider: %v", pName, err)
		return
	}
	//log.Printf("[%s] Found %d possible matches", pName, len(shows))
	wg := sync.WaitGroup{}
	seen := map[string]bool{}

	bar := pc.AddBar(int64(len(shows)),
		mpb.PrependDecorators(
			// simple name decorator
			decor.Name(left(p.Name(), 20), decor.WC{W: 20 + 1, C: decor.DidentRight}),
			decor.CountersNoUnit(" %2d/%2d", decor.WC{W: 5 + 1, C: decor.DidentRight}),
		))
	for _, s := range shows {
		if _, ok := seen[s.ID]; ok {
			continue
		}
		seen[s.ID] = true

		d := a.Config.Destinations[s.Destination]
		if a.Config.Force || a.MustDownload(p, s, d) {
			wg.Add(1)
			a.SubmitDownload(&wg, p, s, d, pc, bar)
		} else {
			bar.Increment()
			//log.Printf("[%s] Show %q, %q is already download", pName, s.Show, s.Title)
		}
	}
	bar.SetTotal(int64(len(shows)), true)
	wg.Wait()
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

func (a *app) SubmitDownload(wg *sync.WaitGroup, p providers.Provider, s *providers.Show, d string, pc *mpb.Progress, bar *mpb.Bar) {
	a.worker.Submit(workers.NewRunAction(fmt.Sprintf("[%s] show: %q", p.Name(), p.GetShowFileName(s)), func() error {
		return a.DownloadShow(wg, p, s, d, pc, bar)
	}))
}

func (a *app) DownloadShow(wg *sync.WaitGroup, p providers.Provider, s *providers.Show, d string, pc *mpb.Progress, bar *mpb.Bar) error {
	done := make(chan bool)
	bar.Increment()

	deleteFile := false
	fn := filepath.Join(d, p.GetShowFileName(s))
	defer func() {
		if deleteFile {
			os.Remove(fn)
		}
		wg.Done()
		close(done)
	}()
	err := os.MkdirAll(filepath.Dir(fn), 0777)
	if err != nil {
		return err
	}

	url, err := p.GetShowStreamURL(s)
	if err != nil {
		return err
	}
	if strings.ToLower(filepath.Ext(url)) == ".m38u" {
		master, err := m3u8.NewMaster(url, a.getter)
		if err != nil {
			return err
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

	cmd := exec.Command("ffmpeg", params...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if !a.Config.Debug {
		go io.Copy(ioutil.Discard, stdout)
		go io.Copy(ioutil.Discard, stderr)
	} else {
		go io.Copy(os.Stdout, stdout)
		go io.Copy(os.Stderr, stderr)
	}

	if a.Config.Debug {
		//log.Printf("[%s] Runing FFMPEG to get %q", p.Name(), p.GetShowFileName(s))
	}

	go func() {
		start := time.Now()

		fileBar := pc.AddSpinner(0, mpb.SpinnerOnMiddle,
			mpb.BarWidth(5),
			mpb.PrependDecorators(
				decor.Name(left(path.Base(p.GetShowFileName(s)), 75), decor.WC{W: 76, C: decor.DidentRight}),
				// decor.OnComplete(decor.Spinner(nil, decor.WCSyncSpace), "done"),
			),
			mpb.AppendDecorators(
				decor.AverageSpeed(decor.UnitKB, " %.1f"),
			),
			mpb.BarRemoveOnComplete(),
		)

		lastSize := 0
		t := time.NewTicker(500 * time.Millisecond)
		f := func() {
			s, err := os.Stat(fn)
			if err != nil {
				return
			}
			l := int(s.Size())
			fileBar.IncrBy(l-lastSize, time.Since(start))
			lastSize = l
		}
		for {
			select {
			case <-t.C:
				f()
			case <-done:
				t.Stop()
				f()
				fileBar.SetTotal(100, true)
				break
			}
		}

	}()

	err = cmd.Run()
	if err != nil {
		deleteFile = true
		return err
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
		return fmt.Errorf("[%s] Can't download %q's thumbnail: %v", p.Name(), p.GetShowFileName(s), err)
	}

	ws := []io.Writer{}
	tbnFile, err := os.Create(tbnFileName)
	if err != nil {
		return fmt.Errorf("[%s] Can't create %q's thumbnail: %v", p.Name(), p.GetShowFileName(s), err)
	}
	defer tbnFile.Close()
	ws = append(ws, tbnFile)

	if mustDownloadShowTbnFile {
		showTbnFile, err := os.Create(showTbnFileName)
		if err != nil {
			return fmt.Errorf("[%s] Can't create shows's %q thumbnail: %v", p.Name(), s.Show, err)
		}
		defer showTbnFile.Close()
		ws = append(ws, showTbnFile)
	}

	wr := io.MultiWriter(ws...)
	_, err = io.Copy(wr, tbnStream)
	if err != nil {
		return fmt.Errorf("[%s] Can't write %q's thumbnail: %v", p.Name(), p.GetShowFileName(s), err)
	}
	if a.Config.Headless {
		log.Printf("%s downloaded.", fn)
	}
	return nil
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
