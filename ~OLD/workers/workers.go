package workers

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/simulot/aspiratv/mylog"
)

// WorkItem is an interface to work item used b the Workers
type WorkItem func()

// Worker is a pool of workers.
type Worker struct {
	stop   chan bool // Close this channel to stop all workers
	tokens chan int
	wg     sync.WaitGroup // To wait completion of all workers
	logger *mylog.MyLog   // True to enable logs
}

// New creates a new worker pool with NumCPU runing workers
func New(ctx context.Context, workers int, logger *mylog.MyLog) *Worker {
	if workers < 1 {
		workers = runtime.NumCPU()
	}
	w := &Worker{
		stop:   make(chan bool),
		tokens: make(chan int, workers),
		logger: logger,
	}

	w.logger.Debug().Printf("[WORKER] Starting %d workers", workers)
	for i := 0; i < workers; i++ {
		w.tokens <- i + 1
	}

	return w
}

// Stop stops all runing workers and wait them to finish and then leave the workerpool.
func (w *Worker) Stop(ctx context.Context) {
	w.logger.Debug().Printf("[WORKER] Stoping...")

	// Makes all goroutine ending
	close(w.stop)

	// wait the end of all job
	w.logger.Debug().Printf("[WORKER] Waiting for worker to end")
	w.wg.Wait()
	w.logger.Debug().Printf("[WORKER] Worker pool is ended")
}

var jobID = int64(0)

// Submit a work item to the worker pool
func (w *Worker) Submit(ctx context.Context, wi func(wid, jid int)) {
	id := atomic.AddInt64(&jobID, 1)
	w.logger.Debug().Printf("[WORKER] Job %d is waiting...", id)
	select {
	case <-w.stop:
		w.logger.Debug().Printf("[WORKER] Stopped, job %d discared", id)
		return
	case <-ctx.Done():
		w.logger.Debug().Printf("[WORKER] Cancled, job %d discared", id)
		return
	case wid := <-w.tokens:
		w.wg.Add(1)
		fn := func(wid, id int) {
			w.logger.Debug().Printf("[WORKER %d] Job %d is started", wid, id)
			wi(wid, id)
			w.logger.Debug().Printf("[WORKER %d] Job %d is done", wid, id)
			w.tokens <- wid
			w.wg.Done()
		}
		go fn(wid, int(id))
	}
}
