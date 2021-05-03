package download

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"time"
)

type (
	logger     interface{ Printf(string, ...interface{}) }
	progresser interface{ Progress(current, total int) }
)

type FFMPEG struct {
	l      logger     // to produce some logs
	p      progresser // to follow progession
	wd     *watchdog  // drop the download after a while
	cmd    *exec.Cmd  // the command
	inputs []string
	output string

	lastOutput strings.Builder
}

func NewFFMPEG() *FFMPEG {
	return &FFMPEG{
		l: log.Default(),
	}
}

func (d *FFMPEG) WithLogger(l interface{ Printf(string, ...interface{}) }) *FFMPEG {
	d.l = l
	return d
}

func (d *FFMPEG) WithProgresser(p progresser) *FFMPEG {
	d.p = p
	return d
}

func (d *FFMPEG) Input(uri string) *FFMPEG {
	d.inputs = append(d.inputs, uri)
	return d
}

func (d *FFMPEG) Download(ctx context.Context, destination string) (err error) {
	d.l.Printf("[FFMPEG] Start download of %s", destination)
	var (
		output io.ReadCloser
	)

	if len(d.inputs) == 0 {
		return errors.New("missing input")
	}
	defer func() {
		if err != nil {
			d.l.Printf("[FFMPEG] ERROR: %s", err.Error())
		}
	}()
	params := []string{
		"-loglevel", "info", // Give me feedback
		"-hide_banner", // I don't want banner
		"-nostdin",
	}
	for _, i := range d.inputs {
		params = append(params, "-i", i)
	}
	params = append(
		params,
		"-vcodec", "copy", // copy video
		"-acodec", "copy", // copy audio
		"-bsf:a", "aac_adtstoasc", // I don't know
		"-y",        // Override output file
		"-f", "mp4", // Be sure that output
		destination, // output file
	)

	d.cmd = exec.CommandContext(ctx, "ffmpeg", params...)
	output, err = d.cmd.StderrPipe()

	if err != nil {
		return err
	}

	d.parseOutput(output)

	d.wd = newWatchDog(60*time.Second, func() {
		d.cmd.Process.Kill()
	})
	err = d.cmd.Start()
	if err != nil {
		return err
	}
	err = d.cmd.Wait()

	d.wd.Stop()
	output.Close()

	if err != nil {
		if d.lastOutput.Len() > 0 {
			err = errors.New(d.lastOutput.String())
		}
	}

	return err
}

func (d *FFMPEG) parseOutput(r io.ReadCloser) {
	sc := bufio.NewScanner(r)
	sc.Split(scanLines)
	go func() {
		total := 0.0
		perCent := 0.0
		estimatedSize := int64(0)

		for sc.Scan() {
			d.wd.Kick()
			l := sc.Bytes()

			if bytes.HasPrefix(l, []byte("frame=")) {
				d.lastOutput.Reset()
				i := bytes.Index(l, []byte("size="))
				if i < 0 {
					continue
				}
				var size int64
				_, err := fmt.Sscanf(string(l[i+len("size="):i+len("size= 1011200")]), "%8d", &size)
				if err != nil {
					continue
				}
				size = size * 1024

				final := bytes.Contains(l, []byte("Lsize="))
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
				if d.p != nil {
					d.p.Progress(int(size), int(estimatedSize))
				}
				continue
			}

			if i := bytes.Index(l, []byte("Duration:")); i >= 0 {
				d.lastOutput.Reset()
				var h, m, s, c int64
				_, err := fmt.Sscanf(string(l[i+len("Duration:"):i+len("Duration: 01:29:25.00")]), "%2d:%2d:%2d.%2d", &h, &m, &s, &c)
				if err != nil {
					continue
				}
				total = float64(h*int64(time.Hour) + m*int64(time.Minute) + s*int64(time.Second) + c*int64(time.Millisecond)/10)
				continue
			}
			if len(l) > 0 && l[0] != ' ' {
				d.lastOutput.Write(l)
				d.lastOutput.WriteString("\r\n")
			}
		}
	}()
}

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
