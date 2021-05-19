package mockup

import (
	"context"
	"log"
	"math/rand"

	"github.com/simulot/aspiratv/pkg/download"

	"github.com/google/uuid"
	"github.com/simulot/aspiratv/pkg/models"
)

func (p Mockup) GetMedias(ctx context.Context, task models.DownloadTask) (<-chan models.DownloadItem, error) {
	show := models.ShowInfo{
		ID:    uuid.NewString(),
		Plot:  "This is the show plot",
		Type:  task.Result.Type,
		Title: task.Result.Show,
		Images: []models.Image{
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
		Images: []models.Image{
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
			Show:       show.Title,
			Title:      "episode title",
			Episode:    1,
			Season:     4,
			SeasonInfo: &season,
			ShowInfo:   &show,
		},
		{
			ID:         uuid.NewString(),
			Show:       show.Title,
			Title:      "episode title",
			Episode:    2,
			Season:     4,
			SeasonInfo: &season,
			ShowInfo:   &show,
		},
		{
			ID:         uuid.NewString(),
			Title:      "episode title",
			Show:       show.Title,
			Episode:    1,
			Season:     3,
			SeasonInfo: &season,
			ShowInfo:   &show,
		},
	}

	c := make(chan models.DownloadItem, 1)

	go func() {
		defer log.Printf("Exit GR mockup.GetMedias")
		for i, m := range list {
			log.Printf("Loop GR mockup.GetMedias %d/%d", i, len(list))
			stream := videoSamples[rand.Intn(len(videoSamples))]
			m.StreamURL = stream
			m.Provider = "mockup"

			item := models.DownloadItem{
				Downloader: download.NewFFMPEG().Input(m.StreamURL),
				MediaInfo:  m,
			}
			log.Printf("Found media: %s", m.ID)
			select {
			case <-ctx.Done():
				log.Printf("GR mockup.GetMedias ctx.done: %s", ctx.Err())
				return
			case c <- item:
				log.Printf("Media send: %s", m.ID)
			}
		}
		log.Printf("End of Loop GR mockup.GetMedias")
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
