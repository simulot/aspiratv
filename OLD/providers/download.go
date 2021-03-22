package providers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"

	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	// Image formats
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/simulot/aspiratv/download"
	"github.com/simulot/aspiratv/media"
	"github.com/simulot/aspiratv/metadata/nfo"
)

type downloader struct {
	r           *Runner
	crumbs                     // Collect created files to be deleted in event of cancellation
	returnedErr error          // The final error
	info        *nfo.MediaInfo // Metadata
	mediaPath   string         // Path of Video
}

func newDownloader(r *Runner) *downloader {
	return &downloader{
		r:      r,
		crumbs: crumbs{},
	}
}

func (d *downloader) done(ctx context.Context) {

	if errors.Is(ctx.Err(), context.Canceled) && d.returnedErr == nil {
		d.returnedErr = ctx.Err()
	}
	if d.returnedErr != nil {
		d.crumbs.cleanFiles()
	}
}

func (d *downloader) download(ctx context.Context, m *media.Media, fb FeedBacker) {
	d.r.c.log.Trace().Printf("[downloader] [%s] Start downloading %q", d.r.p.Name(), m.ShowRootPath)
	defer func() {
		defer d.done(ctx)
		d.r.c.log.Trace().Printf("[downloader] [%s] Exit downloader.download(%s), %s", d.r.p.Name(), m.Metadata.GetMediaInfo().Title, ErrString(d.returnedErr))
	}()
	if fb != nil {
		defer func() {
			fb.Done()
		}()
	}

	d.info = m.Metadata.GetMediaInfo()

	// the media will be downladed into mediaPath file
	d.mediaPath, d.returnedErr = download.MediaPath(m.ShowRootPath, m.Match, d.info)
	if d.returnedErr != nil {
		return
	}

	// create needed dirs where medias will be downloaded
	mediaDir := filepath.Dir(d.mediaPath)
	d.crumbs.addDir(mediaDir)
	d.returnedErr = os.MkdirAll(mediaDir, 0777)
	if d.returnedErr != nil {
		return
	}

	itemName := filepath.Base(d.mediaPath)

	// Get more details from the provider
	d.returnedErr = d.r.p.GetMediaDetails(ctx, m)
	if d.returnedErr != nil {
		return
	}

	url := m.Metadata.GetMediaInfo().MediaURL
	if len(url) == 0 {
		d.returnedErr = fmt.Errorf("Can't get url from %s: empty url", itemName)
		return
	}
	d.r.c.log.Trace().Printf("[downloader] [%s] Player URL for '%s' is %q ", d.r.p.Name(), itemName, url)
	if fb != nil {
		fb.Stage("downloading media...")
		fb.Total(1)
	}
	d.crumbs.addFile(d.mediaPath)

	d.returnedErr = download.Download(ctx, d.r.c.log, url, d.mediaPath, d.info, download.WithLogger(d.r.c.log), download.WithProgress(fb))
	if d.returnedErr != nil {
		return
	}

	// TODO implement DownloadNFO option
	d.writeNFO(ctx, m, fb)

}

func (d *downloader) writeNFO(ctx context.Context, m *media.Media, fb FeedBacker) {
	// fb.Stage("Thumbnails...")

	if d.info.MediaType != nfo.TypeMovie {
		if d.info.TVShow != nil {
			nfoPath := filepath.Join(m.Match.ShowRootPath, "tvshow.nfo")
			nfoExists, err := fileExists(nfoPath)
			if !nfoExists && err == nil {
				d.crumbs.addFile(nfoPath)
				d.info.TVShow.WriteNFO(nfoPath)
				if err != nil {
					d.r.c.log.Error().Printf("WriteNFO: %s", err)
				}
				d.downloadImages(ctx, nfoPath, d.info.TVShow.Thumb)
			}
		}
		if d.info.SeasonInfo != nil {
			seasonPath, err := download.SeasonPath(m.ShowRootPath, m.Match, d.info)
			if err != nil {
				d.r.c.log.Error().Printf("Can't dertermine season path: %s", err)
			}
			nfoPath := filepath.Join(seasonPath, "season.nfo")
			nfoExists, err := fileExists(nfoPath)
			if !nfoExists && err == nil {
				d.addFile(nfoPath)
				d.info.SeasonInfo.WriteNFO(nfoPath)
				if err != nil {
					d.r.c.log.Error().Printf("WriteNFO: %s", err)
				}
				d.downloadImages(ctx, nfoPath, d.info.SeasonInfo.Thumb)
			}
		}
	}

	nfoPath := strings.TrimSuffix(d.mediaPath, filepath.Ext(d.mediaPath)) + ".nfo"
	nfoExists, err := fileExists(nfoPath)
	if !nfoExists && err == nil {
		d.addFile(nfoPath)
		err = m.Metadata.WriteNFO(nfoPath)
		if err != nil {
			d.r.c.log.Error().Printf("WriteNFO: %s", err)
		}
		d.downloadImages(ctx, nfoPath, d.info.Thumb)
	}

}

// downloadImages images listed in nfo.Thumbs into nfoPath. nfo's name  indicates if images are those from episode, season or show.
func (d *downloader) downloadImages(ctx context.Context, nfoPath string, thumbs []nfo.Thumb) {

	nfoDir := filepath.Dir(nfoPath)
	nfoName := filepath.Base(nfoPath)

	for i, thumb := range thumbs {
		var base string
		if errors.Is(ctx.Err(), context.Canceled) {
			return
		}

		switch nfoName {
		case "tvshow.nfo", "season.nfo":
			base = fmt.Sprintf("%s_%d.png", thumb.Aspect, i+1)
		default: // For episodes
			base = strings.TrimSuffix(nfoName, ".nfo")
			if thumb.Aspect == "thumb" || thumb.Aspect == "" {
				base = fmt.Sprintf("%s_%d.png", base, i+1)
			} else {
				continue
			}
		}
		thumbName := filepath.Join(nfoDir, base)
		if thumbExists, _ := fileExists(thumbName); thumbExists {
			continue
		}
		d.downloadImage(ctx, thumb.URL, thumbName)
	}
}

func (d *downloader) downloadImage(ctx context.Context, url, imageName string) {
	resp, err := http.Get(url)
	if err != nil {
		d.r.c.log.Error().Printf("[%s] Can't get thumbnail: %s", err)
		return
	}
	defer resp.Body.Close()

	// /!\ Test the exact type of image ... sometime .jpg is actually .png
	var format string
	buf := bytes.NewBuffer([]byte{})

	tr := io.TeeReader(resp.Body, buf)
	_, format, err = image.DecodeConfig(tr)

	// image with the correct extension
	imageName = strings.TrimSuffix(imageName, filepath.Ext(imageName)) + "." + format
	d.crumbs.addFile(imageName)

	w, err := os.Create(imageName)
	if err != nil {
		d.r.c.log.Error().Printf("[%s] Can't get thumbnail: %s", err)
		return
	}
	defer w.Close()

	mr := io.MultiReader(buf, resp.Body)
	_, err = io.Copy(w, mr)

	if err != nil {
		d.r.c.log.Error().Printf("[%s] Can't get thumbnail: %s", err)
		return
	}
	d.r.c.log.Trace().Printf("[%s] thumbnail %q downloaded.", d.r.p.Name(), imageName)
}

// crumbs collect files and dir names created during the process, to be able to
// delete them in case of cancellation
type crumbs []string

// cleanFiles remove items in reverse order of creation
func (c *crumbs) cleanFiles() {
	for len(*c) > 0 {
		fn := (*c)[len(*c)-1]
		os.Remove(fn)
		*c = (*c)[:len(*c)-1]
	}
}

// addFile add the file into the crumb stack
func (c *crumbs) addFile(fn string) {
	*c = append(*c, fn)
}

// addDir into the crumb stack. Works with dirs with to be created sub dirs
func (c *crumbs) addDir(d string) {
	d, err := filepath.Abs(d)
	if err != nil {
		return
	}
	ps := strings.Split(d, string(os.PathSeparator))

	d = "/"
	for _, p := range ps {
		if p == "" {
			continue
		}
		d = filepath.Join(d, p)
		_, err := os.Stat(d)
		if os.IsNotExist(err) {
			c.addFile(d)
		}
	}

}
