package models

import (
	"time"

	"github.com/google/uuid"
)

type Publishable interface {
	UUID() uuid.UUID
}

type PublishableType int

// Message is a simple message that can be Published
type Message struct {
	ID   uuid.UUID
	When time.Time
	Type PublishableType
	Text string
}

const (
	TypeMessage PublishableType = iota
	TypeNotification
	TypeProgression
)

func NewMessage(t string) Message {
	return Message{
		Type: TypeMessage,
		ID:   uuid.New(),
		Text: t,
	}
}
func (m Message) UUID() uuid.UUID { return m.ID }

// Notification can Publish evolution of an object
type Notification struct {
	ID               uuid.UUID
	Type             PublishableType
	NotificationType NotificationType
	Text             string
}
type NotificationType int

func NewNotification(text string, t NotificationType) Notification {
	return Notification{
		ID:               uuid.New(),
		Type:             TypeNotification,
		NotificationType: t,
		Text:             text,
	}
}
func (n Notification) UUID() uuid.UUID { return n.ID }

const (
	NotificationInfo NotificationType = iota
	NotificationSuccess
	NotificationWarning
	NotificationError
)

// Progression can Publish a progression of a task
type Progression struct {
	ID      uuid.UUID
	Type    PublishableType
	Text    string
	Current float32 // Percent 0-not commenced, 1-finished
}

func NewProgression(t string) Progression {
	return Progression{
		ID:      uuid.New(),
		Type:    TypeProgression,
		Text:    t,
		Current: 0,
	}
}
func (p Progression) UUID() uuid.UUID { return p.ID }
