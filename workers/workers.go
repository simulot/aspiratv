package workers

import (
	"context"
	"log"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/simulot/aspiratv/mylog"
)

// WorkItem is an interface to work item used b the Workers
type WorkItem func()

// WorkerPool is a pool of workers.
type WorkerPool struct {
	stop     chan bool      // Close this channel to stop all workers
	submit   chan WorkItem  // Send work items to this channel, one of workers will run it
	wg       sync.WaitGroup // To wait completion of all workers
	nbWorker int            // The number of concurrent workers
	logger   *mylog.MyLog   // True to enable logs
	ctx      context.Context
}

// New creates a new worker pool with NumCPU runing workers
func New(ctx context.Context, workers int, logger *mylog.MyLog) *WorkerPool {
	if workers < 1 {
		workers = runtime.NumCPU()
	}
	w := &WorkerPool{
		stop:     make(chan bool),
		submit:   make(chan WorkItem),
		nbWorker: workers,
		logger:   logger,
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
		w.logger.Debug().Printf("Worker goroutine %d is shutted down.", index)
	}()
	for {
		select {
		case <-ctx.Done():
			w.logger.Debug().Printf("Worker %d has received a %q", index, ctx.Err())
			return
		case wi := <-w.submit:
			wi()
			w.wg.Done()
		case <-w.stop:
			w.logger.Debug().Printf("Worker %d has received a stop request.", index)
			return
		}
	}
}

// Stop stops all runing workers and wait them to finish and then leave the workerpool.
func (w *WorkerPool) Stop() {
	// wait the end of all job
	w.wg.Wait()
	w.logger.Debug().Printf("Waiting for worker to end")
	// Makes all goroutine ending
	close(w.stop)

	w.logger.Debug().Printf("Worker pool is ended")
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
		w.logger.Debug().Printf("Job %d is started", id)
		w.wg.Add(1)
		wi()
		w.logger.Debug().Printf("Job %d is done", id)

	}:
		w.logger.Debug().Printf("Job %d is queued", id)
		return
	case <-w.ctx.Done():
		w.logger.Debug().Printf("Job %d cancelled", id)
		wg.Done()
		return
	}
}
