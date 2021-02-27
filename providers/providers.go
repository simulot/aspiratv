package providers

import (
	"context"

	"github.com/simulot/aspiratv/matcher"
	"github.com/simulot/aspiratv/media"
	"github.com/simulot/aspiratv/mylog"
	"golang.org/x/time/rate"
)

// Provider is the interface for a provider
type Provider interface {
	Configure(fns ...ProviderConfigFn)                                    // Pass general configuration
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
type ProviderConfig struct {
	Log         *mylog.MyLog  // Logger
	HitsLimiter *rate.Limiter // Limit the number of hits per second
	// TODO ByteLimiter *rate.Limiter // Limit the number of bytes per second
	UserAgent string // User agent to use for queries
}

type ProviderConfigFn func(c ProviderConfig) ProviderConfig

func ProviderLog(l *mylog.MyLog) ProviderConfigFn {
	return func(c ProviderConfig) ProviderConfig {
		c.Log = l
		return c
	}
}

func ProviderHitsPerSecond(limiter *rate.Limiter) ProviderConfigFn {
	return func(c ProviderConfig) ProviderConfig {
		c.HitsLimiter = limiter
		return c
	}
}

func ProviderUserAgent(agent string) ProviderConfigFn {
	return func(c ProviderConfig) ProviderConfig {
		c.UserAgent = agent
		return c
	}
}
