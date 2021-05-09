package dispatcher

import (
	"sync"

	"github.com/simulot/aspiratv/pkg/models"
)

type Subscriber interface {
	// Subscribe call the given function for each new notification.
	// The returned function must be called to cancel the subscription
	Subscribe(func(models.Message)) (cancel func())
}

type Publisher interface {
	// Publish the notification to all current subcribers
	Publish(models.Message)
}

// Dispatcher is in charge of dispatch notifications to
// all its subscribers
type Dispatcher struct {
	sync.RWMutex
	subscribers []*subscriber
}

// subscriber will receive message emitted by the dispatcher
type subscriber struct {
	n chan models.Message
}

// NewDispatcher creates a dispatcher
func NewDispatcher() *Dispatcher {
	d := Dispatcher{}
	return &d
}

// Publish send the notification to all of subscribers
func (d *Dispatcher) Publish(n models.Message) {
	d.RLock()
	defer d.RUnlock()
	for _, s := range d.subscribers {
		s.n <- n
	}
}

// Subscribe call onMessage function for each message and return the Unsubscribe function
// It creates a subcriber record for each subscriber
func (d *Dispatcher) Subscribe(onMessage func(models.Message)) (cancel func()) {
	d.Lock()
	defer d.Unlock()
	s := &subscriber{
		n: make(chan models.Message, 1),
	}
	d.subscribers = append(d.subscribers, s)

	go func() {
		for n := range s.n {
			onMessage(n)
		}
	}()

	return func() {
		d.unsubscribe(s)
	}
}

// unsubscribe remove the subscriber from the list
func (d *Dispatcher) unsubscribe(s *subscriber) {
	d.Lock()
	defer d.Unlock()
	for i := range d.subscribers {
		if d.subscribers[i] == s {
			close(s.n)
			d.subscribers[i] = d.subscribers[len(d.subscribers)-1]
			d.subscribers[len(d.subscribers)-1] = nil
			d.subscribers = d.subscribers[0 : len(d.subscribers)-1]
			return
		}
	}
}
