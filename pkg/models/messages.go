package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type StatusType int

const (
	StatusInfo StatusType = iota
	StatusSuccess
	StatusWarning
	StatusError
)

type PublishableType int

// Message is a simple message that can be Published
type Message struct {
	ID          uuid.UUID           // uuid
	Status      StatusType          // Status Success/Error/Info...
	When        time.Time           // creation / update time
	Text        string              // Textual message
	Progression *ProgressionPayload //
}

func NewMessage(t string, s StatusType) Message {
	return Message{
		Status: s,
		When:   time.Now(),
		ID:     uuid.New(),
		Text:   t,
	}
}

func (m Message) UUID() uuid.UUID { return m.ID }
func (m Message) AutoClose() bool {
	if m.Progression != nil {
		return m.Progression.AutoClose()
	}
	return true
}
func (m Message) String() string {
	if m.Progression != nil {
		return m.Text + " " + m.Progression.String()
	}
	return m.Text
}

// Progression can Publish a progression of a task
type ProgressionPayload struct {
	Current int
	Total   int
}

func NewProgression(t string, s StatusType, current int, total int) Message {
	return Message{
		ID:     uuid.New(),
		Status: s,
		When:   time.Now(),
		Text:   t,
		Progression: &ProgressionPayload{
			Current: current,
			Total:   total,
		},
	}
}
func (p ProgressionPayload) AutoClose() bool {
	return p.Total > p.Current
}

func (p ProgressionPayload) String() string {
	pc := float64(0)
	if p.Total > 0 {
		pc = float64(p.Current) * 100.0 / float64(p.Total)
	}
	return fmt.Sprintf("%3.1f%%", pc)
}
func (p *ProgressionPayload) Progress(current int, total int) {
	p.Current = current
	p.Total = total
}
