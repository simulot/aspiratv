package dispatcher

import (
	"reflect"
	"sync"
	"testing"

	"github.com/simulot/aspiratv/pkg/models"
)

func TestNotifications(t *testing.T) {
	t.Run("Check if a message is received", func(t *testing.T) {
		d := NewDispatcher()
		want := models.Message{
			Text: "This is the text",
		}

		done := make(chan struct{})

		var got *models.Message
		cancel := d.Subscribe(func(p *models.Message) {
			got = p
			close(done)
		})
		d.Publish(&want)
		<-done
		if !reflect.DeepEqual(want, got) {
			t.Errorf("Expecting %v, got %v", want, got)
		}
		cancel()
	})

	t.Run("Check if 10 messages sent to 10 subscribers make 100 receptions", func(t *testing.T) {
		d := NewDispatcher()
		want := models.Message{
			Text: "This is the text",
		}
		var received [10]int

		var wgGRRunning sync.WaitGroup
		var wgReception sync.WaitGroup

		wgGRRunning.Add(10)
		wgReception.Add(10)
		for i := 0; i < 10; i++ {
			go func(i int) {
				var w sync.WaitGroup
				w.Add(10)
				cancel := d.Subscribe(func(p *models.Message) {
					got := p
					if !reflect.DeepEqual(want, got) {
						t.Errorf("Client %d, want %v, got %v", i, want, got)
					}
					received[i]++
					w.Done()
				})
				wgGRRunning.Done()
				w.Wait()
				wgReception.Done()
				cancel()
			}(i)
		}

		wgGRRunning.Wait()
		for i := 0; i < 10; i++ {
			d.Publish(&want)
		}
		wgReception.Wait()

		for i := 0; i < 10; i++ {
			if received[i] != 10 {
				t.Errorf("Expecting subscriber #%d to receive %d, but got %d ", i, 10, received[i])
			}
		}

	})
}
