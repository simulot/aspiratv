package workers

import (
	"log"
	"runtime"
	"strconv"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	w := &WorkerPool{
		stop:     make(chan bool),
		submit:   make(chan WorkItem),
		nbWorker: 1,
		debug:    true,
	}
	w.init()

	for i := 0; i < runtime.NumCPU()*2; i++ {
		w.Submit(
			NewRunAction("test_"+strconv.Itoa(i), func() error {
				log.Println("...")
				time.Sleep(100 * time.Millisecond)
				return nil
			}))
	}
	w.Stop()
}

func TestNew2(t *testing.T) {
	w := &WorkerPool{
		stop:     make(chan bool),
		submit:   make(chan WorkItem),
		nbWorker: 2,
		debug:    true,
	}
	w.init()

	for i := 0; i < runtime.NumCPU()*2; i++ {
		w.Submit(
			NewRunAction("test_"+strconv.Itoa(i), func() error {
				log.Println("...")
				time.Sleep(100 * time.Millisecond)
				return nil
			}))
	}
	w.Stop()
}
