package models

import (
	"time"

	"github.com/google/uuid"
)

//go:generate enumer -type=PoolRhythmType -json
type PoolRhythmType int

const (
	RhythmUnknown PoolRhythmType = iota
	RhythmDaily
	RhythmWeekly
	RhythmMonthly
)

type Subscription struct {
	UUID          uuid.UUID      `json:"uuid,omitempty"`            // Unique identifier
	Enabled       bool           `json:"enabled,omitempty"`         // When true, the query will ran automatically
	Title         string         `json:"title,omitempty"`           // Show title
	ShowPageURL   string         `json:"show_page_url,omitempty"`   // Show landing page, used by the provider to pull new medias
	ShowID        string         `json:"show_id,omitempty"`         // Show ID when available
	ShowType      MediaType      `json:"show_type,omitempty"`       // Determine folder layout
	Provider      string         `json:"provider,omitempty"`        // Provider that host the show
	PollRhythm    PoolRhythmType `json:"poll_rhythm,omitempty"`     // Pooling rhythm
	LastRun       time.Time      `json:"last_run,omitempty"`        // When the query ran for the last time
	LastSeenMedia time.Time      `json:"last_seen_media,omitempty"` // When the query has returned a  media for the last time
	LimitNumber   int            `json:"limit_number,omitempty"`    // Limit number of media to be downloaded
	MaxAge        int            `json:"max_age,omitempty"`         // When set, the media must be aired after Now()-MaxAge days
	DeleteAge     int            `json:"delete_age,omitempty"`      // When set, medias are deleted when aired after Now()-DeleteAge days
}
