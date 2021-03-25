package store

import "errors"

type Store interface {
	ProviderInterface
}

type ProviderInterface interface {
	GetProvider(name string) (Provider, error)
	SetProvider(p Provider) (Provider, error)
	GetProviderList() ([]Provider, error)
}

// Provider structure used by the API
type Provider struct {
	Name string // Provider handle name
}

/*
type SubscriptionInterface interface {
	GetSubscription(id string) (Subscription, error)
	SetSubscription(s Subscription) (Subscription, error)
	GetSubciptionList() ([]Subscription, error)
}

type Subscription struct {
}
*/

var ErrorNotFound = errors.New("ressource not found")
