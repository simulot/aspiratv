package models

import (
	"sort"
	"time"

	"github.com/google/uuid"
)

type Runner interface {
	Run()
}
type TaskType int

const (
	Unknown TaskType = iota
	SubscriptionPooling
	// TBD: Library scan
)

type Task struct {
	UUID   uuid.UUID // Task ID
	When   time.Time // When the action must be run
	Type   TaskType  // Task type
	Title  string    //  Title for user interface
	Runner Runner    // Runner to be invoked when time is arrived
}

type Schedule []Task

func (s Schedule) SortedByTime() Schedule {
	sorted := s
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].When.After(sorted[j].When)
	})
	return sorted
}
