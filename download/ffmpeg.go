package download

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"time"

	"github.com/simulot/aspiratv/metadata/nfo"
)

func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data

}
func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, dropCR(data[0:i]), nil
	}
	if i := bytes.IndexByte(data, '\r'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, dropCR(data[0:i]), nil
	} // If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropCR(data), nil
	}
	// Request more data.
	return 0, nil, nil
}

func (c *ffmpegConfig) watchProgress(r io.ReadCloser, prg Progresser) {
	sc := bufio.NewScanner(r)
	sc.Split(scanLines)
	go func() {
		const (
			start int = iota
			inInput
			inRunning
		)
		total := 0.0
		perCent := 0.0
		estimatedSize := int64(0)
		var lastLine []byte

		wf := 10 * time.Second
		if c.conf.debug {
			wf = 10 * time.Hour
		}

		// watch if frames are comming
		activityWatchDog := newWatchDog(wf, func() {
			lastLine = []byte("time out when receiving frames")
			c.cmd.Process.Kill()
		})
		defer activityWatchDog.Stop()

		for sc.Scan() {
			if c.conf.debug && len(lastLine) > 0 {
				if !bytes.HasPrefix(lastLine, []byte("frame=")) {
					log.Print("[FFMPEG] ", string(lastLine))
				}
			}

			l := sc.Bytes()
			lastLine = l

			if bytes.HasPrefix(l, []byte("frame=")) {

				//  alive!
				activityWatchDog.Kick()

				i := bytes.Index(l, []byte("size="))
				if i < 0 {
					continue
				}
				var size int64
				_, err := fmt.Sscanf(string(l[i+len("size="):i+len("size= 1011200")]), "%8d", &size)
				size = size * 1024

				final := bytes.Index(l, []byte("Lsize=")) >= 0

				i = bytes.Index(l, []byte("time="))
				if i < 0 {
					continue
				}
				var h, m, s, c int64
				_, err = fmt.Sscanf(string(l[i+len("time="):i+len("time=00:41:52.49")]), "%2d:%2d:%2d.%2d", &h, &m, &s, &c)
				if err != nil {
					continue
				}
				if final {
					estimatedSize = size
				} else {
					current := h*int64(time.Hour) + m*int64(time.Minute) + s*int64(time.Second) + c*int64(time.Millisecond)/10
					perCent = float64(current) / total
					estimatedSize = int64(float64(size) / perCent)
					if estimatedSize < size {
						estimatedSize = size + 1024
					}
				}
				if prg != nil {
					prg.Update(size, estimatedSize)
				}
				continue
			}

			if i := bytes.Index(l, []byte("Duration:")); i >= 0 {
				var h, m, s, c int64
				_, err := fmt.Sscanf(string(l[i+len("Duration:"):i+len("Duration: 01:29:25.00")]), "%2d:%2d:%2d.%2d", &h, &m, &s, &c)
				if err != nil {
					continue
				}
				total = float64(h*int64(time.Hour) + m*int64(time.Minute) + s*int64(time.Second) + c*int64(time.Millisecond)/10)
				if prg != nil {
					prg.Init(int64(1 * 1024 * 1024))
				}
			}

		}
		if !bytes.HasPrefix(lastLine, []byte("frame=")) {
			log.Print("[FFMPEG] ", string(lastLine))
		}
		c.lastLine = string(lastLine)
	}()
}

type ffmpegConfig struct {
	conf     *DownloadConfiguration
	lastLine string
	cmd      *exec.Cmd
}

func FFMpeg(ctx context.Context, in, out string, info *nfo.MediaInfo, configurations ...ConfigurationFunction) error {
	cfg := ffmpegConfig{
		conf: NewDownloadConfiguration(),
	}
	for _, c := range configurations {
		c(cfg.conf)
	}
	params := []string{}
	if len(cfg.conf.params) == 0 {
		params = []string{
			"-loglevel", "info", // Give me feedback
			"-hide_banner", // I don't want banner
			"-nostdin",
			"-i", in, // Where is the stream
			"-vcodec", "copy", // copy video
			"-acodec", "copy", // copy audio
			"-bsf:a", "aac_adtstoasc", // I don't know
			"-metadata", "title=" + info.Title, // Force title
			"-metadata", "comment=" + info.Plot, // Force comment
			"-metadata", "show=" + info.Showtitle, //Force show
			"-metadata", "channel=" + info.Studio, // Force channel
			"-y",        // Override output file
			"-f", "mp4", // Be sure that output
			out, // output file
		}
	}
	if cfg.conf.debug {
		log.Printf("[FFMPEG] running ffmpeg %v", params)
	}

	cfg.cmd = exec.CommandContext(ctx, "ffmpeg", params...)
	stdOut, err := cfg.cmd.StderrPipe()

	cfg.watchProgress(stdOut, cfg.conf.pgr)
	err = cfg.cmd.Start()
	if err != nil {
		return err
	}
	err = cfg.cmd.Wait()
	if err != nil {
		err = fmt.Errorf("FFMPEG can't process stream %s,\n %w", cfg.lastLine, err)
	}

	return err
}
