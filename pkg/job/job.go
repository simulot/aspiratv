package job

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

// A job launch a series of tasks
type Job struct {
	concurrency int // Number of concurrent jobs
	tokens      chan struct{}
	name        string
}

// a task is a work unit
type Task func() error

type JobNotificationType int

const (
	JobStarted JobNotificationType = iota
	JobTaskStarted
	JobTaskRunning
	JobTaskSuccess
	JobTaskError
	JobEnded
)

type JobNotification struct {
	Type    JobNotificationType
	JobUUID uuid.UUID
	TaskID  uuid.UUID
	Message string
}

func NewJob(name string) *Job {
	return &Job{
		name:        name,
		concurrency: 1,
	}
}

func (j *Job) Run(ctx context.Context, tasks <-chan Task) {
	j.tokens = make(chan struct{}, j.concurrency)
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
					select {
					case <-ctx.Done():
						return
					case task, ok := <-tasks:
						if !ok {
							return
						}
						task()
						j.tokens <- struct{}{}
					}
				}
			}
		}(i)
	}
	wg.Wait()
}
