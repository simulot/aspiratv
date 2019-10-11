package main

import (
	"context"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

func (a *app) DownloadToPNG(ctx context.Context, url, imageName string) error {
	imageName = strings.TrimSuffix(imageName, filepath.Ext(imageName)) + ".png"
	thumbStream, err := a.getter.Get(ctx, url)
	if err != nil {
		return err
	}

	w, err := os.Create(imageName)
	if err != nil {
		return err
	}
	defer w.Close()
	defer thumbStream.Close()

	img, _, err := image.Decode(thumbStream)
	if err != nil {
		return err
	}
	e := png.Encoder{
		CompressionLevel: png.BestCompression,
	}
	return e.Encode(w, img)
}
