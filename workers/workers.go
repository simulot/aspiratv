package workers

import (
	"log"
	"runtime"
	"sync"
	"time"
)

type WorkItem interface {
	Run() error
	Name() string
}

type WorkerPool struct {
	stop    chan bool
	submit  chan WorkItem
	workerg sync.WaitGroup
}

func New() *WorkerPool {
	w := &WorkerPool{
		stop:   make(chan bool),
		submit: make(chan WorkItem),
	}
	w.Init(runtime.NumCPU())
	return w
}

func (w *WorkerPool) Submit(wi WorkItem) {
	w.submit <- wi
}
func (w *WorkerPool) Init(n int) *WorkerPool {
	w.workerg.Add(n)
	for i := 0; i < n; i++ {
		go w.NewWorker(i)
	}
	return w
}

func (w *WorkerPool) NewWorker(id int) {
	for {
		select {
		case <-w.stop:
			w.workerg.Done()
			return
		case i := <-w.submit:
			// log.Printf("Start [%d]: %s\n", id, i.Name())
			t := time.Now()
			err := i.Run()
			if err == nil {
				log.Printf("Done  [%d]: (%s) %s\n", id, time.Since(t), i.Name())
			} else {
				log.Printf("Fail  [%d]: %s\n", id, i.Name())
				log.Printf("Error [%d]: %v\n", id, err)
			}
		}
	}
}

type RunAction struct {
	name string
	fn   func() error
}

func NewRunAction(n string, fn func() error) RunAction {
	return RunAction{name: n, fn: fn}
}
func (r RunAction) Name() string {
	return r.name
}
func (r RunAction) Run() error {
	return r.fn()
}
