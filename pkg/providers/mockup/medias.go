package mockup

import (
	"context"
	"math/rand"

	"github.com/google/uuid"
	"github.com/simulot/aspiratv/pkg/models"
)

func (p Mockup) GetMedias(ctx context.Context, task models.DownloadTask) (<-chan models.DownloadItem, error) {
	show := models.ShowInfo{
		ID:    uuid.NewString(),
		Plot:  "This is the show plot",
		Type:  task.Result.Type,
		Title: task.Result.Show,
		Thumbs: []models.Image{
			{
				ID:     uuid.NewString(),
				Aspect: "banner",
				URL:    "https://via.placeholder.com/1920x1080.png",
			},
		},
	}

	season := models.SeasonInfo{
		ID:   uuid.NewString(),
		Plot: "This is the season plot",
		Thumbs: []models.Image{
			{
				ID:     uuid.NewString(),
				Aspect: "banner",
				URL:    "https://via.placeholder.com/1920x1080.png",
			},
		},
	}

	list := []models.MediaInfo{
		{
			ID:         uuid.NewString(),
			Title:      "title",
			Episode:    1,
			Season:     4,
			SeasonInfo: &season,
			ShowInfo:   &show,
		},
		{
			ID:         uuid.NewString(),
			Title:      "title",
			Episode:    2,
			Season:     4,
			SeasonInfo: &season,
			ShowInfo:   &show,
		},
		{
			ID:         uuid.NewString(),
			Title:      "title",
			Episode:    1,
			Season:     3,
			SeasonInfo: &season,
			ShowInfo:   &show,
		},
	}

	c := make(chan models.DownloadItem)

	go func() {
		for _, m := range list {
			item := models.DownloadItem{
				Downloader: models.DownloaderFFMPEG,
				MediaInfo:  m,
			}
			item.MediaInfo.StreamURL = videoSamples[rand.Intn(len(videoSamples))]
			select {
			case <-ctx.Done():
				return
			case c <- item:
			}
		}
		close(c)
	}()

	return c, nil
}

var videoSamples = []string{
	"http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4",
	"http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ElephantsDream.mp4",
	"http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ForBiggerBlazes.mp4",
	"http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ForBiggerEscapes.mp4",
	"http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ForBiggerFun.mp4",
	"http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ForBiggerJoyrides.mp4",
	"http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ForBiggerMeltdowns.mp4",
	"http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/Sintel.mp4",
	"http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/SubaruOutbackOnStreetAndDirt.mp4",
	"http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/TearsOfSteel.mp4",
	"http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/VolkswagenGTIReview.mp4",
	"http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/WeAreGoingOnBullrun.mp4",
	"http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/WhatCarCanYouGetForAGrand.mp4",
}
