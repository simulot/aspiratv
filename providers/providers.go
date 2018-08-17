package providers

import (
	"encoding/gob"
)

type Provider interface {
	Name() string                           // Provider's name
	Shows() ([]*Show, error)                // List of available shows
	GetShowStreamURL(*Show) (string, error) // Give video stream url ofr a give show
	GetShowFileName(*Show) string           // Give a sensible name for the given show
	GetShowFileNameMatcher(*Show) string    // Give a file name matcher for searching duplicates having different episode number
}

type Configurer interface {
	SetConfig(map[string]interface{})
}

var providers = map[string]Provider{}

func Register(p Provider) {
	providers[p.Name()] = p
}

func List() map[string]Provider {
	return providers
}

func init() {
	gob.Register([]*Show{})
}
