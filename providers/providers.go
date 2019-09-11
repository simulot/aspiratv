package providers

import (
	"encoding/gob"
)

// Provider is the interface for a provider
type Provider interface {
	Name() string                           // Provider's name
	Shows([]*MatchRequest) chan *Show       // List of available shows that match one of MatchRequest
	GetShowStreamURL(*Show) (string, error) // Give video stream url ofr a give show
	GetShowFileName(*Show) string           // Give a sensible name for the given show
	GetShowFileNameMatcher(*Show) string    // Give a file name matcher for searching duplicates having different episode number
}

// Configurer interface
type Configurer interface {
	SetConfig(map[string]interface{})
}

var providers = map[string]Provider{}

// Register is called by provider's init to register the provider
func Register(p Provider) {
	providers[p.Name()] = p
}

// List of registered providers
func List() map[string]Provider {
	return providers
}

func init() {
	gob.Register([]*Show{})
}
