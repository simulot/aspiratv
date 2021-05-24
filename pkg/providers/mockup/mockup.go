/*
	This providers is used to have a test bed for building the application without querying actual web sites
*/

package mockup

import (
	"context"
	"fmt"
	"sync"
	"time"

	"math/rand"

	"github.com/google/uuid"
	"github.com/simulot/aspiratv/pkg/models"
	"github.com/simulot/aspiratv/pkg/providers"
)

type Mockup struct {
}

var (
	channels = providers.Description{
		Code: "mockup",
		Name: "mMockupTV",
		Logo: "/web/tv.svg",
		Channels: map[string]providers.Channel{
			"mockup": {
				Code: "mockup",
				Name: "Mockup Channel",
				Logo: "/web/arte.png",
			},
		},
	}
)

func NewMockup() *Mockup {
	return &Mockup{}
}

func (Mockup) Name() string { return "mockup" }

func (Mockup) ProviderDescribe(ctx context.Context) providers.Description {
	return channels
}

func (p *Mockup) Search(ctx context.Context, q models.SearchQuery) (<-chan models.SearchResult, error) {
	results := make(chan models.SearchResult, 1)

	go p.callSearch(ctx, results, q)
	return results, nil
}

var initSeed sync.Once

func (p *Mockup) callSearch(ctx context.Context, results chan models.SearchResult, q models.SearchQuery) {
	defer close(results)

	initSeed.Do(func() {
		rand.Seed(time.Now().Unix())
	})

	count := 1 + rand.Intn(2)
	i := 1

	for count >= 0 {
		t := time.After(time.Millisecond * time.Duration(200+rand.Intn(1000)))

		r := models.SearchResult{
			ID:              uuid.NewString(),
			Type:            models.MediaType(rand.Intn(int(models.TypeMovie) + 1)),
			Chanel:          "mockup",
			Provider:        "mockup",
			Show:            "The " + q.Title + " show",
			Title:           fmt.Sprintf("%q Item %d", q.Title, i),
			ThumbURL:        "https://via.placeholder.com/1920x1080.png",
			PageURL:         "https://www.youtube.com/watch?v=x31tDT-4fQw",
			AvailableVideos: rand.Intn(7) + 1,
			Plot: `
			Generated Lorem Ipsum
			
			Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. 
			Dignissim sodales ut eu sem integer vitae justo eget magna. Rhoncus mattis rhoncus urna neque viverra justo nec ultrices. 
			Nec nam aliquam sem et. Id eu nisl nunc mi ipsum faucibus. Massa tincidunt dui ut ornare lectus sit amet est placerat. 
			Placerat vestibulum lectus mauris ultrices. Arcu cursus vitae congue mauris rhoncus. Auctor eu augue ut lectus. 
			Faucibus in ornare quam viverra orci sagittis eu volutpat. In eu mi bibendum neque egestas. Mi eget mauris pharetra et ultrices. 
			Tempus urna et pharetra pharetra massa massa ultricies. Augue lacus viverra vitae congue eu consequat. Purus viverra accumsan in nisl.`,
		}
		select {
		case <-ctx.Done():
			return
		case <-t:
			results <- r
		}
		count--
		i++
	}
}
