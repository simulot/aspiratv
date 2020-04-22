package download

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/simulot/aspiratv/metadata/nfo"
)

type Progresser interface {
	Init(size int64)
	Update(count int64, size int64)
}

type DownloadConfiguration struct {
	debug  bool
	pgr    Progresser
	params map[string]string
}

func NewDownloadConfiguration() *DownloadConfiguration {
	return &DownloadConfiguration{
		params: map[string]string{},
	}
}

type ConfigurationFunction func(*DownloadConfiguration)

func Download(ctx context.Context, in, out string, info *nfo.MediaInfo, conf ...ConfigurationFunction) error {

	req, err := http.NewRequest("HEAD", in, nil)
	if err != nil {
		return err
	}

	transport := http.Transport{}
	resp, err := transport.RoundTrip(req)
	if err != nil {
		return err
	}

	// // Check the content header
	// resp, err := http.Head(in)
	// if err != nil {
	// 	return err
	// }

	if l := resp.Header.Get("Location"); l != "" {
		// Substitute in URL by relocation address
		in = l
		resp, err = http.Head(in)
		if err != nil {
			return err
		}
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP Error: %s", err)
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
		return DASH(ctx, in, out, info, conf...)

	case "FFMPEG":
		return FFMpeg(ctx, in, out, info, conf...)
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

func WithDebug(debug bool) ConfigurationFunction {
	return func(c *DownloadConfiguration) {
		c.debug = debug
	}
}

func WithParams(params map[string]string) ConfigurationFunction {
	return func(c *DownloadConfiguration) {
		c.params = params
	}
}
