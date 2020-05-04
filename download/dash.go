package download

import (
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
)

type DASHConfig struct {
	getTokens chan bool
	conf      *DownloadConfiguration
	mpd       *mpdparser.MPDParser
	bytesRead int64
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
	err := d.mpd.Get(ctx, in)
	if err != nil {
		return fmt.Errorf("[DASH] Can't get manifest: %s", err)
	}

	videoIT, err := d.progression(in, "video/mp4")
	if err != nil {
		return err
	}

	audioIT, err := d.getSegments(in, "audio/mp4")

	if err != nil {
		return err
	}
	var returnedErr error

	defer func() {
		if err != nil {
			log.Error().Printf("[DASH] %w", returnedErr)
		} else {
			log.Trace().Printf("[DASH] successful download of %s", out)
		}
		os.Remove(out + ".audio.mp4")
		os.Remove(out + ".video.mp4")
		if returnedErr != nil {
			os.Remove(out)
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(2)
	var errVideo, errAudio error

	go func() {
		err := d.downloadSegments(ctx, out+".video.mp4", videoIT)
		if err != nil {
			errVideo = fmt.Errorf("Video segment: %w", err)
			cancel()
		}
		wg.Done()
	}()
	go func() {
		err := d.downloadSegments(ctx, out+".audio.mp4", audioIT)
		if err != nil {
			errAudio = err
			errVideo = fmt.Errorf("Audio segment: %w", err)
			cancel()
		}
		wg.Done()
	}()

	wg.Wait()
	switch {
	case errVideo != nil:
		returnedErr = errVideo
	case errAudio != nil:
		returnedErr = errAudio
	}

	if returnedErr != nil {
		return returnedErr
	}

	// Combine the streams
	params := []string{
		"-i", out + ".video.mp4",
		"-i", out + ".audio.mp4",
		"-codec", "copy",
	}
	if info != nil {
		params = append(params,
			"-metadata", "title="+info.Title, // Force title
			"-metadata", "comment="+info.Plot, // Force comment
			"-metadata", "show="+info.Showtitle, //Force show
			"-metadata", "channel="+info.Studio, // Force channel
		)
	}
	params = append(params,
		"-f", "mp4",
		"-y",
		out,
	)
	cmd := exec.Command("ffmpeg", params...)
	returnedErr = cmd.Run()
	return returnedErr
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

func (d *DASHConfig) progression(manifest string, mime string) (mpdparser.SegmentIterator, error) {
	it, err := d.getSegments(manifest, mime)
	if err != nil {
		return nil, err
	}
	return &progressionIterator{
		d:  d,
		it: it,
	}, nil
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

func (d *DASHConfig) downloadSegments(ctx context.Context, filename string, it mpdparser.SegmentIterator) error {
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
			n, err := io.CopyBuffer(f, r.Body, nil)
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
