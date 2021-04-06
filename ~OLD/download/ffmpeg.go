package download

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
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

func (c *ffmpegConfig) watchProgress(r io.ReadCloser, fb FeedBacker) {
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
		var lastLine []byte // Keep the last line which contains the real error

		// watch if frames are comming
		activityWatchDog := newWatchDog(60*time.Second, func() {
			lastLine = []byte("time out when receiving frames")
			c.cmd.Process.Kill()
		})
		defer activityWatchDog.Stop()

		for sc.Scan() {
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
				if fb != nil {
					fb.Total(int(estimatedSize))
					fb.Update(int(size))
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
				if fb != nil {
					fb.Total(int(1 * 1024 * 1024))
				}
			}

		}
		c.lastLine = string(lastLine)
	}()
}

type ffmpegConfig struct {
	conf     *downloadConfiguration
	lastLine string
	cmd      *exec.Cmd
}

func ffmpeg(ctx context.Context, in, out string, info *nfo.MediaInfo, configurations ...configurationFunction) error {
	cfg := ffmpegConfig{
		conf: newDownloadConfiguration(),
	}
	for _, c := range configurations {
		c(cfg.conf)
	}
	defer func() {
		if cfg.conf.fb != nil {
			cfg.conf.fb.Done()
		}
		cfg.conf.logger.Trace().Printf("[FFMPEG] Exit, %s,%s", out, ctx.Err())
	}()
	params := []string{
		"-loglevel", "info", // Give me feedback
		"-hide_banner", // I don't want banner
		"-nostdin",
		"-i", in, // Where is the stream
		"-vcodec", "copy", // copy video
		"-acodec", "copy", // copy audio
		"-bsf:a", "aac_adtstoasc", // I don't know
		// "-metadata", "title=" + info.Title, // Force title
		// "-metadata", "comment=" + info.Plot, // Force comment
		// "-metadata", "show=" + info.Showtitle, //Force show
		// "-metadata", "channel=" + info.Studio, // Force channel
		"-y",        // Override output file
		"-f", "mp4", // Be sure that output
		out, // output file
	}
	cfg.conf.logger.Trace().Printf("[FFMPEG] running ffmpeg %v", params)

	cfg.cmd = exec.CommandContext(ctx, "ffmpeg", params...)
	stdOut, err := cfg.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("[FFMPEG] %w", err)
	}

	cfg.watchProgress(stdOut, cfg.conf.fb)
	err = cfg.cmd.Start()
	if err != nil {
		return err
	}
	err = cfg.cmd.Wait()
	if err != nil {
		err = fmt.Errorf("[FFMPEG] Error %s,\n %w", cfg.lastLine, err)
	}

	return err
}
