package providers

import (
	"context"
)

// Provider is the interface for a provider
type Provider interface {
	Name() string                                           // Provider's name
	MediaList(context.Context, []*MatchRequest) chan *Media // List of available shows that match one of MatchRequest
	GetMediaDetails(context.Context, *Media) error          // Download more details when available
	DebugMode(bool)                                         // Set debug mode
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
