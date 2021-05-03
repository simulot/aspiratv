package frontend

import (
	"fmt"
	"time"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/simulot/aspiratv/pkg/dispatcher"
	"github.com/simulot/aspiratv/pkg/frontend/bulma"
	"github.com/simulot/aspiratv/pkg/models"
)

// NotificationsDrawer collect published events and keep them
// until the user has dismissed them. This struct lives into
// to global state of the application in order to survives to
// page navigation
type NotificationsDrawer struct {
	p           []models.Publishable
	subscribers []chan struct{}
}

// NewNotificationsDrawer create a new drawer
func NewNotificationsDrawer() *NotificationsDrawer {
	d := NotificationsDrawer{}
	return &d
}

// Attach a notification provider to the drawer
func (d *NotificationsDrawer) Attach(sub dispatcher.Subscriber) {
	sub.Subscribe(d.onNotification)
}

// onNotification is the called back when a new publication arrives
// when the incoming publication is already known ( same UUID) the new one
// replace the old one. Then drawer subscribers are notified
func (d *NotificationsDrawer) onNotification(new models.Publishable) {
	defer d.notify()
	for i := range d.p {
		if d.p[i].UUID() == new.UUID() {
			d.p[i] = new
			return
		}
	}
	d.p = append(d.p, new)
}

// notify dispatch notification to all subscribers
func (d *NotificationsDrawer) notify() {
	for _, c := range d.subscribers {
		c <- struct{}{}
	}
}

// OnChange attach subscriber's call back function each time
// notify is called  (when a new publication is arrived)
// it return a function to be called to unsubscribe
func (d *NotificationsDrawer) OnChange(fn func()) func() {
	c := make(chan struct{}, 1)
	go func() {
		for range c {
			fn()
		}
	}()
	d.subscribers = append(d.subscribers, c)

	// cancel function
	return func() {
		for i := 0; i < len(d.subscribers); i++ {
			if d.subscribers[i] == c {
				close(c)
				d.subscribers[i] = d.subscribers[len(d.subscribers)-1]
				d.subscribers[len(d.subscribers)-1] = nil
				d.subscribers = d.subscribers[0 : len(d.subscribers)-1]
			}
		}
	}
}

// Notifications get the list of all notifications maintained in the drawer
func (d *NotificationsDrawer) Notifications() []models.Publishable {
	r := []models.Publishable{}
	for i := 0; i < len(d.p); i++ {
		r = append(r, d.p[i])
	}
	return r
}

// Dismiss remove a notification when the user click on close or a timeout run off
func (d *NotificationsDrawer) Dismiss(p models.Publishable) {
	id := p.UUID()
	for i := range d.p {
		if d.p[i].UUID() == id {
			copy(d.p[i:], d.p[i+1:])                // Shift d.n[i+1:] left one index.
			d.p[len(d.p)-1] = models.Notification{} // Erase last element (write zero value).
			d.p = d.p[:len(d.p)-1]                  // Truncate slice.
			d.notify()
			return
		}
	}
}

// PublishableContainer display pending notifications
type PublishableContainer struct {
	app.Compo
	unsubscribe func()
}

func NewPublishableContainer() *PublishableContainer {
	return &PublishableContainer{}
}

func (c *PublishableContainer) OnMount(ctx app.Context) {
	c.unsubscribe = MyAppState.Drawer.OnChange(func() {
		ctx.Dispatch(func(ctx app.Context) {
			// no op
		})
	})
}

func (c *PublishableContainer) OnDismount() {
	if c.unsubscribe != nil {
		c.unsubscribe()
	}
}

func (c *PublishableContainer) Render() app.UI {
	ns := MyAppState.Drawer.Notifications()
	return app.Div().
		Class("toast-container column is-4 is-offset-8").
		Body(
			app.Range(ns).
				Slice(func(i int) app.UI {
					return NewPublishableElement(ns[i])
				}),
		)
}

type PublishableElement struct {
	app.Compo
	Dismiss func()
	models.Publishable
}

func NewPublishableElement(p models.Publishable) *PublishableElement {
	c := &PublishableElement{
		Dismiss:     func() { MyAppState.Drawer.Dismiss(p) },
		Publishable: p,
	}
	time.AfterFunc(4*time.Second, c.Dismiss)
	return c
}

func (c *PublishableElement) Render() app.UI {

	switch p := c.Publishable.(type) {
	case models.Notification:
		return c.RenderNotification(p)
	case models.Message:
		return c.RenderMessage(p)
	default:
		return c.RenderMessage(models.NewMessage(fmt.Sprintf("Unknown notification type: %T", p)))
	}
}

func (c *PublishableElement) RenderMessage(m models.Message) app.UI {
	return bulma.NewNotification().
		Class("is-info").
		Text(m.Text)
}
func (c *PublishableElement) RenderNotification(m models.Notification) app.UI {
	class := map[models.NotificationType]string{
		models.NotificationError:   "is-danger",
		models.NotificationInfo:    "is-info",
		models.NotificationSuccess: "is-success",
		models.NotificationWarning: "is-warning",
	}[m.NotificationType]

	n := bulma.NewNotification().Class(class).Text(m.Text)
	if m.NotificationType == models.NotificationError {
		n.OnDelete(c.Dismiss)
	}
	return n
}
