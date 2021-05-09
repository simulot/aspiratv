package download

import (
	"time"
)

type watchdog struct {
	interval time.Duration
	timer    *time.Timer
}

// newWatchDog
func newWatchDog(interval time.Duration, callback func()) *watchdog {
	w := watchdog{
		interval: interval,
		timer:    time.AfterFunc(interval, callback),
	}
	return &w
}

func (w *watchdog) Stop() {
	w.timer.Stop()
}

func (w *watchdog) Kick() {
	w.timer.Stop()
	w.timer.Reset(w.interval)
}
