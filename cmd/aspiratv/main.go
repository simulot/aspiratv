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
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	_ "github.com/simulot/aspiratv/providers/artetv"
	_ "github.com/simulot/aspiratv/providers/francetv"

	"github.com/simulot/aspiratv/net/http"
	"github.com/simulot/aspiratv/playlists/m3u8"
	"github.com/simulot/aspiratv/providers"

	"github.com/simulot/aspiratv/workers"
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
	flag.StringVar(&cliConfig.ConfigFile, "config", "config.json", "Configuration file name.")
	flag.Parse()

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
		log.Fatal("Missing ffmpeg on your system, it is required to handle video files.")
	}
	a.ffmpeg = strings.Trim(strings.Trim(string(b), "\r\n"), "\n")
	if a.Config.Debug {
		log.Printf("FFMPEG path: %q", a.ffmpeg)
	}

}

func (a *app) RunOnce() {
	// Kick of providers, wait queries to finish, and exit
	wg := sync.WaitGroup{}
	for _, p := range providers.List() {
		wg.Add(1)
		go func(p providers.Provider) {
			a.PullShows(p)
			wg.Done()
		}(p)
	}
	wg.Wait()
	log.Println("Job(s) are done!")
}

func (a *app) RunAsService() {
	// Kick of providers loop and remain active
	for n, p := range providers.List() {
		go a.ProviderLoop(p)
		log.Printf("Provider %s watch loop initialized", n)
	}
	<-a.Stop
}

func (a *app) ProviderLoop(p providers.Provider) {
	for {
		select {
		case <-a.Stop:
			return
		default:
			a.PullShows(p)
			s := time.Duration(a.Config.PullInterval) + time.Duration(rand.Intn(int(time.Duration(a.Config.PullInterval))/4))
			log.Printf("[%s] Provider is sleeping until %s\n", p.Name(), time.Now().Add(s).Format("15:04:05"))
			time.Sleep(s)
		}
	}
}

type pullWork struct {
	worker      *workers.WorkerPool
	config      *Config
	wg          sync.WaitGroup
	getter      getter
	ffmpeg      string
	deduplicate map[string]bool
}

type debugger interface {
	SetDebug(bool)
}

func (a *app) PullShows(p providers.Provider) {
	w := &pullWork{
		worker:      workers.New(),
		config:      a.Config,
		getter:      http.DefaultClient,
		ffmpeg:      a.ffmpeg,
		deduplicate: map[string]bool{},
	}
	if d, ok := p.(debugger); ok {
		d.SetDebug(a.Config.Debug)
	}
	w.Run(p)
}

func (w *pullWork) Run(p providers.Provider) {
	pName := p.Name()
	shows, err := p.Shows(w.config.WatchList)
	if err != nil {
		log.Printf("[%s] Can't get shows list of provider: %v", pName, err)
		return
	}
	log.Printf("[%s] %s has %d show(s) that match", pName, pName, len(shows))
	for _, s := range shows {
		d := w.config.Destinations[s.Destination]
		if w.config.Force || w.MustDownload(p, s, d) {
			w.wg.Add(1)
			w.SubmitDownload(p, s, d)
		} else {
			if w.config.Debug {
				log.Printf("[%s] Show %q, %q is already download", pName, s.Show, s.Title)
			}
		}
	}
	w.wg.Wait()
	w.worker.Stop()
}

func (w *pullWork) MustDownload(p providers.Provider, s *providers.Show, d string) bool {
	if _, ok := w.deduplicate[s.ID]; ok {
		return false
	}
	w.deduplicate[s.ID] = true
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

func (w *pullWork) SubmitDownload(p providers.Provider, s *providers.Show, d string) {
	w.worker.Submit(workers.NewRunAction(fmt.Sprintf("[%s] Downloading show: %q", p.Name(), p.GetShowFileName(s)), func() error {
		return w.DownloadShow(p, s, d)
	}))
}

func (w *pullWork) DownloadShow(p providers.Provider, s *providers.Show, d string) error {

	deleteFile := false
	fn := filepath.Join(d, p.GetShowFileName(s))
	defer func() {
		if deleteFile {
			os.Remove(fn)
		}
		w.wg.Done()
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
		master, err := m3u8.NewMaster(url, w.getter)
		if err != nil {
			return err
		}
		url = master.BestQuality()
	}

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

	if !w.config.Debug {
		go io.Copy(ioutil.Discard, stdout)
		go io.Copy(ioutil.Discard, stderr)
	} else {
		go io.Copy(os.Stdout, stdout)
		go io.Copy(os.Stderr, stderr)
	}

	if w.config.Debug {
		log.Printf("[%s] Runing FFMPEG to get %q", p.Name(), p.GetShowFileName(s))
	}
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

	tbnStream, err := w.getter.Get(s.ThumbnailURL)
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
	return nil
}

type app struct {
	Config *Config
	Stop   chan bool
	ffmpeg string
}

type getter interface {
	Get(uri string) (io.Reader, error)
}
