package models

import (
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type Subscriber interface {
	// Subscribe call the given function for each new notification.
	// The returned function must be called to cancel the subscription
	Subscribe(func(Notification)) (cancel func())
}

type Publisher interface {
	// Publish the notification to all current subcribers
	Publish(n Notification)
}

type NotificationType int

const (
	NotificationInfo NotificationType = iota
	NotificationSuccess
	NotificationWarning
	NotificationError
)

// Notification carries notification accross modules
type Notification struct {
	id   int32
	when time.Time
	Type NotificationType
	Text string
	// Payload future use: like information on latest downloaded media
}

func (n Notification) ID() int         { return int(n.id) }
func (n Notification) When() time.Time { return n.when }

// NotificationDispatcher is in charge of dispatch notifications to
// all its subscribers
type NotificationDispatcher struct {
	sync.RWMutex
	subscribers []*subscriber
	lastID      int32
}

// subscriber will receive message emitted by the dispatcher
type subscriber struct {
	n chan Notification
}

// NewNotificationDispatcher creates a dispatcher
func NewNotificationDispatcher() *NotificationDispatcher {
	d := NotificationDispatcher{}
	return &d
}

// Publish send the notification to all of subscribers
func (d *NotificationDispatcher) Publish(n Notification) {
	n.id = atomic.AddInt32(&d.lastID, 1)
	n.when = time.Now()
	log.Printf("Publish %v", n)
	d.RLock()
	defer d.RUnlock()
	for _, s := range d.subscribers {
		s.n <- n
	}
}

// Subscribe call onMessage function for each message and return the Unsubscribe function
// It creates a subcriber record for each subscriber
func (d *NotificationDispatcher) Subscribe(onMessage func(Notification)) (cancel func()) {
	d.Lock()
	defer d.Unlock()
	s := &subscriber{
		n: make(chan Notification, 1),
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
func (d *NotificationDispatcher) unsubscribe(s *subscriber) {
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
