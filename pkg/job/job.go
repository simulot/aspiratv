package job

import (
	"context"
	"log"
	"sync"
)

// A job launch a series of tasks
type Job struct {
	concurrency int // Number of concurrent jobs
	tokens      chan struct{}
	running     sync.WaitGroup
}

// a task is a work unit
type Task func() error

func NewJob() *Job {
	return &Job{
		concurrency: 1,
	}
}

func (j *Job) End() {
	close(j.tokens)
	j.running.Wait()
}

func (j *Job) Run(ctx context.Context, tasks <-chan Task) {
	j.running.Add(1)
	defer j.running.Done()
	j.tokens = make(chan struct{}, j.concurrency)
	for i := 0; i < j.concurrency; i++ {
		j.tokens <- struct{}{}
	}
	wg := sync.WaitGroup{}
	wg.Add(j.concurrency)
	for i := 0; i < j.concurrency; i++ {
		go func(runner int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case <-j.tokens:
					log.Printf("Got a job token")
					select {
					case <-ctx.Done():
						return
					case task, ok := <-tasks:
						log.Printf("Go ta task, %v", ok)
						if !ok {
							return
						}
						task()
						log.Printf("Task done")
						j.tokens <- struct{}{}
						log.Printf("Token returned")

					}
				}
			}
		}(i)
	}
	wg.Wait()
}
