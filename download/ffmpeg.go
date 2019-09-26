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

func watchProgress(r io.ReadCloser, prg Progresser) {
	sc := bufio.NewScanner(r)
	go func() {
		const (
			start int = iota
			inInput
			inRunning
		)
		total := 0.0
		perCent := 0.0
		estimatedSize := int64(0)

		state := start
		for sc.Scan() {
			l := sc.Bytes()

			if state == inRunning {
				if !bytes.HasPrefix(l, []byte("frame=")) {
					continue
				}

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

			if i := bytes.Index(l, []byte("Input #")); i >= 0 {
				state = inInput
				continue
			}

			if i := bytes.Index(l, []byte("Press [q] to stop")); i >= 0 {
				state = inRunning
				continue
			}

			if state == inInput {
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

		}
	}()
}

type ffmpegConfig struct {
	debug bool
	pgr   Progresser
}

type ffmpegConfigurator func(c *ffmpegConfig)

func FFMepg(ctx context.Context, u string, params []string, configurators ...ffmpegConfigurator) error {
	cfg := ffmpegConfig{}

	for _, c := range configurators {
		c(&cfg)
	}

	if cfg.debug {
		log.Printf("[FFMPEG] runing ffmpeg %v", params)
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", params...)
	out, err := cmd.StderrPipe()
	if cfg.pgr != nil {
		watchProgress(out, cfg.pgr)
	}
	err = cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()

	return err
}

func FFMepgWithProgress(pgr Progresser) ffmpegConfigurator {
	return func(c *ffmpegConfig) {
		c.pgr = pgr
	}
}

func FFMepgWithDebug(debug bool) ffmpegConfigurator {
	return func(c *ffmpegConfig) {
		c.debug = debug
	}
}
