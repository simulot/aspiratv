package workers

import (
	"context"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
)

// WorkItem is an interface to work item used b the Workers
type WorkItem func()

// WorkerPool is a pool of workers.
type WorkerPool struct {
	stop     chan bool      // Close this channel to stop all workers
	submit   chan WorkItem  // Send work items to this channel, one of workers will run it
	wg       sync.WaitGroup // To wait completion of all workers
	nbWorker int            // The number of concurrent workers
	debug    bool           // True to enable logs
	ctx      context.Context
}

// New creates a new worker pool with NumCPU runing workers
func New(ctx context.Context, workers int, debug bool) *WorkerPool {
	if workers < 1 {
		workers = runtime.NumCPU()
	}
	w := &WorkerPool{
		stop:     make(chan bool),
		submit:   make(chan WorkItem),
		nbWorker: workers,
		debug:    debug,
		ctx:      ctx,
	}
	for i := 0; i < workers; i++ {
		go w.run(ctx, i)
	}
	return w
}

// init creates a goroutine for worker
func (w *WorkerPool) run(ctx context.Context, index int) {
	defer func() {
		if w.debug {
			log.Printf("Worker goroutine %d is shutted down.", index)
		}
	}()
	for {
		select {
		case <-ctx.Done():
			if w.debug {
				log.Printf("Worker %d has recieved a %q", index, ctx.Err())
			}
			return
		case wi := <-w.submit:
			wi()
			w.wg.Done()
		case <-w.stop:
			if w.debug {
				log.Printf("Worker %d has recieved a stop request.", index)
			}
			return
		}
	}
}

// Stop stops all runing workers and wait them to finish and then leave the workerpool.
func (w *WorkerPool) Stop() {
	// wait the end of all job
	w.wg.Wait()
	// Makes all goroutine ending
	close(w.stop)
	if w.debug {
		log.Print("Waiting for worker to end")
	}
	if w.debug {
		log.Print("Worker is ended")
	}
}

var jobID = int64(0)

// Submit a work item to the worker pool
func (w *WorkerPool) Submit(wi WorkItem, wg *sync.WaitGroup) {

	id := atomic.AddInt64(&jobID, 1)
	select {
	case w.submit <- func() {
		defer wg.Done()
		if w.ctx.Err() != nil {
			log.Printf("Job %d is discarded", id)
			return
		}
		if w.debug {
			log.Printf("Job %d is started", id)
		}
		w.wg.Add(1)
		wi()
		if w.debug {
			log.Printf("Job %d is done", id)
		}

	}:
		if w.debug {
			log.Printf("Job %d is queued", id)
		}
		return
	case <-w.ctx.Done():
		if w.debug {
			log.Printf("Job %d cancelled", id)
		}
		wg.Done()
		return
	}
}
