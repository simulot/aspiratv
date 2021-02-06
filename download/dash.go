package download

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"

	"github.com/simulot/aspiratv/metadata/nfo"
	"github.com/simulot/aspiratv/mylog"
	"github.com/simulot/aspiratv/parsers/mpdparser"
	"github.com/simulot/aspiratv/parsers/ttml"
)

type DASHConfig struct {
	getTokens     chan bool
	conf          *DownloadConfiguration
	mpd           *mpdparser.MPDParser
	bytesRead     int64
	lastFFMPGLine string
	cmd           *exec.Cmd
}

// DASH download mp4 file at media's url.
// Open DSH manifest to get best audio and video streams.
// Then download both streams and combine them using FFMPEG
func DASH(ctx context.Context, log *mylog.MyLog, in, out string, info *nfo.MediaInfo, conf ...ConfigurationFunction) error {
	ctx, cancel := context.WithCancel(ctx)

	const concurentDownloads = 2

	defer cancel()

	d := &DASHConfig{
		conf:      NewDownloadConfiguration(),
		getTokens: make(chan bool, concurentDownloads), // concurentDownloads chunks at a time
	}

	// Give tokens for start
	for i := 0; i < concurentDownloads; i++ {
		d.getTokens <- true
	}

	// Apply configuration functions
	for _, c := range conf {
		c(d.conf)
	}

	d.mpd = mpdparser.NewMPDParser()
	d.conf.logger.Trace().Printf("[DASH] Get manifest at %q", in)
	err := d.mpd.Get(ctx, in)
	if err != nil {
		return fmt.Errorf("[DASH] Can't get manifest: %s", err)
	}

	segmentIterators := []mpdparser.SegmentIterator{}
	segmentFileName := []string{}

	for _, as := range d.mpd.Period[0].AdaptationSet {

		best := as.GetBestRepresentation()
		if best == nil {
			continue
		}

		it, err := d.mpd.MediaURIs(in, d.mpd.Period[0], as, best)
		if err != nil {
			return fmt.Errorf("Can't get segments list: %s", err)
		}
		d.conf.logger.Trace().Printf("[DASH] Found representation for type=%q, lang=%q, representation=%q", as.ContentType, as.Lang, best.ID)
		segmentIterators = append(segmentIterators, it)
		segmentFileName = append(segmentFileName, out+"."+it.Content()+"-"+it.Lang()+".mp4")
	}

	var returnedErr error

	defer func() {
		if returnedErr != nil {
			log.Error().Printf("[DASH] %v", returnedErr)
		} else {
			log.Trace().Printf("[DASH] successful download of %s", out)
		}
		for _, k := range segmentFileName {
			os.Remove(k)
		}
	}()

	wg := sync.WaitGroup{}
	for k, it := range segmentIterators {
		wg.Add(1)
		go func(k int, it mpdparser.SegmentIterator) {
			switch it.Content() {
			case "text":
				returnedErr = d.downloadSegments(ctx, segmentFileName[k], it, ttml.TrancodeToSRT)
			case "video":
				returnedErr = d.downloadSegments(ctx, segmentFileName[k], d.progression(it), straitCopy)
			default:
				returnedErr = d.downloadSegments(ctx, segmentFileName[k], it, straitCopy)
			}
			if err != nil {
				cancel()
			}
			wg.Done()
		}(k, it)
	}
	wg.Wait()

	if returnedErr != nil {
		return returnedErr
	}

	// Combine the streams and subtiles
	// http://zoid.cc/12/12/ffmpeg-audio-video/
	// https://en.wikibooks.org/wiki/FFMPEG_An_Intermediate_Guide/subtitle_options
	params := []string{}
	for _, f := range segmentFileName {
		params = append(params, "-i", f)
	}

	for i, it := range segmentIterators {
		switch it.Content() {
		case "audio":
			params = append(params, "-map", fmt.Sprintf("%d:a", i))
		case "video":
			params = append(params, "-map", fmt.Sprintf("%d:v", i))
		case "text":
			params = append(params, "-map", fmt.Sprintf("%d:s", i))
		}
	}

	for i, it := range segmentIterators {
		lang := "eng"
		switch it.Lang() {
		case "fr":
			lang = "fra"
		case "de":
			lang = "deu"
		case "nl":
			lang = "dut"
		case "it":
			lang = "ita"
		case "sp":
			lang = "spa"
		case "da":
			lang = "dan"
		}

		switch it.Content() {
		case "audio":
			params = append(params, fmt.Sprintf("-metadata:s:%d", i), fmt.Sprintf("language=%s", lang))
		case "text":
			params = append(params, fmt.Sprintf("-metadata:s:%d", i), fmt.Sprintf("language=%s", lang), fmt.Sprintf("-metadata:s:%d", i), fmt.Sprintf("title=Subtitles %s", lang))
		}
	}

	params = append(params, "-c:a", "copy")
	params = append(params, "-c:v", "copy")
	params = append(params, "-c:s", "mov_text")

	// if info != nil {
	// 	params = append(params,
	// 		"-metadata", "title="+info.Title, // Force title
	// 		"-metadata", "comment="+info.Plot, // Force comment
	// 		"-metadata", "show="+info.Showtitle, //Force show
	// 		"-metadata", "channel="+info.Studio, // Force channel
	// 	)
	// }
	params = append(params,
		"-f", "mp4",
		"-y",
		out,
	)

	d.conf.logger.Trace().Printf("[DASH] ffmpeg %q", params)

	d.cmd = exec.CommandContext(ctx, "ffmpeg", params...)
	stdOut, returnedErr := d.cmd.StderrPipe()
	if returnedErr != nil {
		return fmt.Errorf("[DASH] %w", returnedErr)
	}
	d.watchFFMPG(stdOut)

	returnedErr = d.cmd.Start()
	if returnedErr != nil {
		return fmt.Errorf("[DASH] %w", returnedErr)
	}

	returnedErr = d.cmd.Wait()
	if returnedErr != nil {
		returnedErr = fmt.Errorf("[FFMPEG] Error %s,\n %w", d.lastFFMPGLine, returnedErr)
	}

	return returnedErr
}

func (d *DASHConfig) watchFFMPG(r io.Reader) {

	sc := bufio.NewScanner(r)
	sc.Split(scanLines)
	go func() {
		const (
			start int = iota
			inInput
			inRunning
		)
		var lastLine []byte // Keep the last line which contains the real error
		for sc.Scan() {
			l := sc.Bytes()
			lastLine = l
		}
		d.lastFFMPGLine = string(lastLine)
	}()
}

func (d *DASHConfig) getSegments(manifest, mime string) (mpdparser.SegmentIterator, error) {

	as := d.mpd.Period[0].GetAdaptationSetByMimeType(mime)
	if as == nil {
		return nil, fmt.Errorf("Missing adaption set for '%s'", mime)
	}
	best := as.GetBestRepresentation()
	if best == nil {
		return nil, fmt.Errorf("Missing Representation for '%s'", mime)
	}

	it, err := d.mpd.MediaURIs(manifest, d.mpd.Period[0], as, best)

	if err != nil {
		return nil, fmt.Errorf("Can't get segments list: %s", err)
	}
	return it, nil
}

type progressionIterator struct {
	d  *DASHConfig
	it mpdparser.SegmentIterator
}

func (d *DASHConfig) progression(it mpdparser.SegmentIterator) mpdparser.SegmentIterator {
	return &progressionIterator{
		d:  d,
		it: it,
	}
}

func (p *progressionIterator) Content() string {
	return p.it.Content()
}
func (p *progressionIterator) Lang() string {
	return p.it.Lang()
}

func (p *progressionIterator) Cancel() {
	p.it.Cancel()
}

func (p *progressionIterator) Next() <-chan mpdparser.SegmentItem {
	c := make(chan mpdparser.SegmentItem)
	go func() {
		for s := range p.it.Next() {
			if p.d.conf.pgr != nil {
				if s.Position.Duration > 0 && s.Position.Time > 0 {
					read := atomic.LoadInt64(&p.d.bytesRead)
					percent := float64(s.Position.Time) / float64(s.Position.Duration)
					estimated := int64(float64(read) / percent)
					if estimated < read {
						estimated = read + 1024
					}
					p.d.conf.pgr.Update(read, estimated)
				}
			}
			c <- s
		}
		p.d.conf.pgr.Update(p.d.bytesRead, p.d.bytesRead)
		close(c)
	}()
	return c
}

func (p *progressionIterator) Err() error {
	return p.it.Err()
}

type tFilter func(dst io.Writer, src io.Reader) (written int64, err error)

func straitCopy(dst io.Writer, src io.Reader) (written int64, err error) {
	return io.CopyBuffer(dst, src, nil)
}

func (d *DASHConfig) downloadSegments(ctx context.Context, filename string, it mpdparser.SegmentIterator, filter tFilter) error {
	ctx, cancel := context.WithCancel(ctx)

	cancelled := false
	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer func() {
		f.Close()
		if cancelled {
			os.Remove(filename)
		}
		cancel()
	}()

	for s := range it.Next() {
		<-d.getTokens
		select {
		case <-ctx.Done():
			d.getTokens <- true
			cancelled = true
			return ctx.Err()
		default:
			if s.Err != nil {
				d.getTokens <- true
				cancelled = true
				return fmt.Errorf("Can't get segment: %w", s.Err)
			}
			d.conf.logger.Debug().Printf("[DASH] Get segment %q", s.S)
			r, err := http.Get(s.S)
			if err != nil {
				d.getTokens <- true
				cancelled = true
				it.Cancel()
				return fmt.Errorf("Can't get segment: %w", err)
			}
			if r.StatusCode >= 400 {
				it.Cancel()
				d.getTokens <- true
				cancelled = true
				return fmt.Errorf("Can't get segment: %s", r.Status)
			}
			// n, err := io.CopyBuffer(f, r.Body, nil)
			n, err := filter(f, r.Body)
			if err != nil {
				r.Body.Close()
				cancelled = true
				d.getTokens <- true
				it.Cancel()
				return fmt.Errorf("Can't add segment: %w", err)
			}
			r.Body.Close()
			atomic.AddInt64(&d.bytesRead, n)

		}
		d.getTokens <- true
	}
	return nil
}
