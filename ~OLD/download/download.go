package download

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/simulot/aspiratv/metadata/nfo"
	"github.com/simulot/aspiratv/mylog"
)

// FeedBacker is an iterface for providing feed back on concurrent tasks
// like a progression bar on the console
type FeedBacker interface {
	Stage(stage string) // Indicate current stage
	Total(total int)    // Indicate the total number (could be bytes, percent )
	Update(current int) // Indicate the current position
	Done()              // Call when the task is done
}

type downloadConfiguration struct {
	fb     FeedBacker
	logger *mylog.MyLog
	// params map[string]string
}

func newDownloadConfiguration() *downloadConfiguration {
	return &downloadConfiguration{
		// params: map[string]string{},
	}
}

type configurationFunction func(*downloadConfiguration)

// Download determine the type of media at given url and launch the appropriate download method
func Download(ctx context.Context, log *mylog.MyLog, in, out string, info *nfo.MediaInfo, configfn ...configurationFunction) error {
	req, err := http.NewRequest("HEAD", in, nil)
	if err != nil {
		return fmt.Errorf("Download: can't get HEAD, %w", err)
	}

	transport := http.Transport{}
	resp, err := transport.RoundTrip(req)
	if err != nil {
		return fmt.Errorf("Download: can't get HEAD, %w", err)
	}

	if l := resp.Header.Get("Location"); l != "" {
		// Substitute in URL by relocation address
		in = l
		resp, err = http.Head(in)
		if err != nil {
			return err
		}
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("[DOWNLOAD] HTTP Error: %s", err)
	}

	downloader := ""
	switch resp.Header.Get("content-type") {
	case "application/dash+xml":
		downloader = "DASH"
	// case "application/vnd.apple.mpegurl":
	// 	downloader = "HLS"
	case "video/mp4":
		downloader = "FFMPEG"
	default:
		url, err := url.Parse(in)
		if err != nil {
			return err
		}
		switch strings.ToLower(path.Ext(path.Base(url.Path))) {
		case ".mp4":
			downloader = "FFMPEG"
		case ".mpd":
			downloader = "DASH"
		// case ".m3u":
		// 	downloader = "HLS"
		default:
			downloader = "FFMPEG"
		}
	}

	switch downloader {
	case "DASH":
		return DASH(ctx, in, out, info, configfn...)

	case "FFMPEG":
		return ffmpeg(ctx, in, out, info, configfn...)
	}
	if downloader == "" {
		return fmt.Errorf("How to download this:%s", in)
	}
	return nil
}

// WithProgress add a feedbacker to the configuration
func WithProgress(fb FeedBacker) configurationFunction {
	return func(c *downloadConfiguration) {
		c.fb = fb
	}
}

// WithLogger add a logger to the configuration
func WithLogger(logger *mylog.MyLog) configurationFunction {
	return func(c *downloadConfiguration) {
		c.logger = logger
	}
}
