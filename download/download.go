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

type Progresser interface {
	Init(size int64)
	Update(count int64, size int64)
}

type DownloadConfiguration struct {
	pgr    Progresser
	logger *mylog.MyLog
	// params map[string]string
}

func NewDownloadConfiguration() *DownloadConfiguration {
	return &DownloadConfiguration{
		// params: map[string]string{},
	}
}

type ConfigurationFunction func(*DownloadConfiguration)

// Download determine the type of media at given url and launch the appropriate download method
func Download(ctx context.Context, log *mylog.MyLog, in, out string, info *nfo.MediaInfo, configfn ...ConfigurationFunction) error {
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
		return DASH(ctx, log, in, out, info, configfn...)

	case "FFMPEG":
		return FFMpeg(ctx, log, in, out, info, configfn...)
	}
	if downloader == "" {
		return fmt.Errorf("How to download this:%s", in)
	}
	return nil
}

func WithProgress(pgr Progresser) ConfigurationFunction {
	return func(c *DownloadConfiguration) {
		c.pgr = pgr
	}
}

func WithLogger(logger *mylog.MyLog) ConfigurationFunction {
	return func(c *DownloadConfiguration) {
		c.logger = logger
	}
}

// func WithDebug(debug bool) ConfigurationFunction {
// 	return func(c *DownloadConfiguration) {
// 		c.debug = debug
// 	}
// }

// func WithParams(params map[string]string) ConfigurationFunction {
// 	return func(c *DownloadConfiguration) {
// 		c.params = params
// 	}
// }
