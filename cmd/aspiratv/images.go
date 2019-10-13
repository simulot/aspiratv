package main

import (
	"bytes"
	"context"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func (a *app) DownloadImage(ctx context.Context, url, imageName string, downloadedFiles *[]string) error {
	thumbStream, err := a.getter.Get(ctx, url)
	if err != nil {
		return err
	}
	defer thumbStream.Close()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		var format string
		var err error
		buf := bytes.NewBuffer([]byte{})

		tr := io.TeeReader(thumbStream, buf)
		_, format, err = image.DecodeConfig(tr)

		imageName = strings.TrimSuffix(imageName, filepath.Ext(imageName)) + ".tmp"
		defer func() {
			*downloadedFiles = append(*downloadedFiles, imageName)
		}()

		w, err := os.Create(imageName)
		if err != nil {
			return err
		}
		defer w.Close()

		mr := io.MultiReader(buf, thumbStream)
		_, err = io.Copy(w, mr)

		if err != nil {
			return err
		}

		new := strings.TrimSuffix(imageName, filepath.Ext(imageName)) + "." + format
		err = os.Rename(imageName, new)
		if err != nil {
			return err
		}
		imageName = new
		return err
	}
}
