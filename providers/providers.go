package providers

import (
	"context"

	"github.com/simulot/aspiratv/matcher"
	"github.com/simulot/aspiratv/media"
	"github.com/simulot/aspiratv/mylog"
)

// Provider is the interface for a provider
type Provider interface {
	Configure(c Config)                                                   // Pass general configuration
	Name() string                                                         // Provider's name
	MediaList(context.Context, []*matcher.MatchRequest) chan *media.Media // List of available shows that match one of MatchRequest
	GetMediaDetails(context.Context, *media.Media) error                  // Download more details when available
}

// register of imported providers
var providers = map[string]Provider{}

// Register is called by provider's init to register the provider
func Register(p Provider) {
	providers[p.Name()] = p
}

// List of registered providers
func List() map[string]Provider {
	return providers
}

// Config carries the configuration to providers
type Config struct {
	Log       *mylog.MyLog // Logger
	KeepBonus bool         // Flag
	// PreferredAudionLang    []string
	// PreferredSubtitlesLang []string
}
