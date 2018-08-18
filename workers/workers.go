package workers

import (
	"log"
	"runtime"
	"sync"
	"time"
)

// WorkItem is an interface to work item used b the Workers
type WorkItem interface {
	Run() error
	Name() string
}

// WorkerPool is a pool of workers.
type WorkerPool struct {
	stop     chan bool      // Close this channel to stop all workers
	submit   chan WorkItem  // Send work items to this channel, one of workers will run it
	workerg  sync.WaitGroup // To wait completion of all workers
	nbWorker int            // The number of concurrent workers
	debug    bool           // True to enable logs
}

// New creates a new worker pool with NumCPU runing workers
func New() *WorkerPool {
	w := &WorkerPool{
		stop:     make(chan bool),
		submit:   make(chan WorkItem),
		nbWorker: runtime.NumCPU(),
	}
	w.init()
	return w
}

// init creates a goroutine for each worker
func (w *WorkerPool) init() *WorkerPool {
	for i := 0; i < w.nbWorker; i++ {
		w.workerg.Add(1)
		go w.newWorker(i)
	}
	return w
}

// Stop stops all runing workers and wait them to finish and then leave the workerpool.
func (w *WorkerPool) Stop() {
	close(w.stop)
	w.workerg.Wait()
	if w.debug {
		log.Print("Workerpool is ended")
	}
}

// Submit a work item to the worker pool
func (w *WorkerPool) Submit(wi WorkItem) {
	if w.debug {
		log.Printf("Submit work:%s", wi.Name())
	}
	w.submit <- wi
}

// newWorker initializes a worker
func (w *WorkerPool) newWorker(id int) {
	if w.debug {
		log.Printf("Initializing worker %d", id)
	}
	for {
		select {
		case <-w.stop:
			w.workerg.Done()
			if w.debug {
				log.Printf("Worker %d is ended", id)
			}
			return
		case i := <-w.submit:
			// log.Printf("Start [%d]: %s\n", id, i.Name())
			t := time.Now()
			err := i.Run()
			if err == nil {
				log.Printf("Done  [%d]: %s(%s)\n", id, i.Name(), time.Since(t).Round(100*time.Millisecond))
			} else {
				log.Printf("Fail  [%d]: %s with error(%v)\n", id, i.Name(), err)
			}
		}
	}
}

// RunAction is an helper to submit a work to the worker pool
type RunAction struct {
	name string
	fn   func() error
}

// NewRunAction creates a work item out of a name and a function
func NewRunAction(n string, fn func() error) RunAction {
	return RunAction{name: n, fn: fn}
}

// Name returns the names of the work
func (r RunAction) Name() string {
	return r.name
}

// Run invoke the function
func (r RunAction) Run() error {
	return r.fn()
}
