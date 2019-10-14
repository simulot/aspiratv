package providers

import (
	"context"
)

// Provider is the interface for a provider
type Provider interface {
	Configure(c Config)                                     // Pass general configuration
	Name() string                                           // Provider's name
	MediaList(context.Context, []*MatchRequest) chan *Media // List of available shows that match one of MatchRequest
	GetMediaDetails(context.Context, *Media) error          // Download more details when available
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

type Config struct {
	Debug     bool
	KeepBonus bool
}
