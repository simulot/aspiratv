package library

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/simulot/aspiratv/pkg/dispatcher"
	"github.com/simulot/aspiratv/pkg/download"
	"github.com/simulot/aspiratv/pkg/models"
)

type (
	Publisher interface {
		Publish(*models.Message)
	}

	Logger interface {
		Printf(string, ...interface{})
	}
)

type BatchDownloader struct {
	name        string
	libraryPath string
	fileNamer   *models.FileNamer
	p           Publisher
	l           Logger
}

func NewBatchDownloader(name string, path string, fn *models.FileNamer) *BatchDownloader {
	return &BatchDownloader{
		name:        name,
		libraryPath: path,
		fileNamer:   fn,
		l:           nullLogger{},
	}
}

type nullLogger struct{}

func (nullLogger) Printf(string, ...interface{}) {}

func (b *BatchDownloader) WithPublisher(p Publisher) *BatchDownloader {
	b.p = p
	return b
}

func (b *BatchDownloader) WithLogger(l Logger) *BatchDownloader {
	b.l = l
	return b
}

func (b *BatchDownloader) Download(ctx context.Context, medias <-chan models.DownloadItem) {
	b.l.Printf("[LIBRARY DOWNLOAD] Start of download %s", b.name)
	defer b.l.Printf("[LIBRARY DOWNLOAD] End of download %s", b.name)

	dispatchError := func(err error) {
		b.l.Printf("[LIBRARY DOWNLOAD] Erreur while downloading %s: %s", b.name, err)
		m := models.NewMessage(fmt.Sprintf("Erreur lors du téléchargement de %q: %s", b.name, err)).SetPinned(true).SetStatus(models.StatusError)
		if b.p != nil {
			b.p.Publish(m)
		}
	}

	libraryPath := download.PathClean(b.libraryPath)

	for {
		select {

		case <-ctx.Done():
			dispatchError(ctx.Err())
			return

		case item, ok := <-medias:
			if !ok {
				return
			}
			t := NewBatch()
			showPath, err := b.fileNamer.ShowPath(item.MediaInfo)
			if err != nil {
				dispatchError(err)
				return
			}
			err = t.Do(MkDirAll(filepath.Join(libraryPath, showPath)))
			if err != nil {
				dispatchError(err)
				return
			}
			if item.MediaInfo.ShowInfo != nil {
				buf, err := json.MarshalIndent(item.MediaInfo.ShowInfo, "", "  ")
				if err != nil {
					dispatchError(err)
					return
				}
				err = t.Do(WriteFile(filepath.Join(libraryPath, showPath, "show.json"), buf))
				if err != nil {
					dispatchError(err)
					return
				}
				downloadImages(ctx, t, filepath.Join(libraryPath, showPath), "", item.MediaInfo.ShowInfo.Images)
			}

			seasonPath, err := b.fileNamer.SeasonPath(item.MediaInfo)
			if err != nil {
				dispatchError(err)
				return
			}
			err = t.Do(MkDirAll(filepath.Join(libraryPath, showPath, seasonPath)))
			if err != nil {
				dispatchError(err)
				return
			}
			if item.MediaInfo.SeasonInfo != nil {
				buf, err := json.MarshalIndent(item.MediaInfo.SeasonInfo, "", "  ")
				if err != nil {
					dispatchError(err)
					return
				}
				err = t.Do(WriteFile(filepath.Join(libraryPath, showPath, seasonPath, "season.json"), buf))
				if err != nil {
					dispatchError(err)
					return
				}
			}

			mediaFileName, err := b.fileNamer.MediaFileName(item.MediaInfo)
			if err != nil {
				dispatchError(err)
				return
			}
			buf, err := json.MarshalIndent(item.MediaInfo, "", "  ")
			if err != nil {
				dispatchError(err)
				return
			}
			err = t.Do(WriteFile(filepath.Join(libraryPath, showPath, seasonPath, strings.TrimSuffix(mediaFileName, filepath.Ext(mediaFileName))+".json"), buf))
			if err != nil {
				dispatchError(err)
				return
			}
			mediaFile := filepath.Join(libraryPath, showPath, seasonPath, mediaFileName)
			downloadImages(ctx, t, filepath.Join(libraryPath, showPath, seasonPath), mediaFileName, item.MediaInfo.Images)

			if !fileExist(mediaFile) {
				itemProgression := dlProgression{
					Message: models.NewProgression(mediaFileName, 0, 0).SetPinned(true).SetStatus(models.StatusInfo),
					p:       b.p,
				}
				if b.p != nil {
					b.p.Publish(itemProgression.Message)
					item.Downloader.WithProgresser(&itemProgression)
				}
				err = t.Do(NewAction(fmt.Sprintf("Download %q", mediaFile), func() error {
					return item.Downloader.Download(ctx, mediaFile)
				}).WithUndo(func() error {
					return os.Remove(mediaFile)
				}))
				if err != nil {
					dispatchError(err)
				}
			}
		}
	}
}

type dlProgression struct {
	*models.Message
	p dispatcher.Publisher
}

func (p *dlProgression) Progress(current int, total int) {
	p.Message.Progression.Progress(current, total)
	if current >= total {
		p.Status = models.StatusSuccess
		p.SetPinned(false)
	}
	p.p.Publish(p.Message)
}

func downloadImages(ctx context.Context, t *Batch, path string, mediaName string, images []models.Image) error {
	extMedia := filepath.Ext(mediaName)
	baseMedia := strings.TrimSuffix(mediaName, extMedia)
	for _, aspect := range []string{"thumb", "backdrop", "poster"} {
		idx := 1
		for _, image := range images {
			if image.Aspect == aspect {
				var imageName string
				if len(mediaName) > 0 {
					switch aspect {
					case "thumb":
						imageName = fmt.Sprintf("%s-%d.png", baseMedia, idx)
					default:
						imageName = fmt.Sprintf("%s-%s-%d.png", baseMedia, aspect, idx)
					}
				} else {
					imageName = fmt.Sprintf("%s-%d.png", aspect, idx)
				}
				err := t.Do(DownloadImage(ctx, image.URL, filepath.Join(path, imageName)))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
