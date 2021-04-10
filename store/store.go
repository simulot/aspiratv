package store

import (
	"context"
	"errors"
)

type Store interface {
	ProviderInterface
}

type ProviderInterface interface {
	GetProvider(ctx context.Context, name string) (Provider, error)
	SetProvider(ctx context.Context, p Provider) (Provider, error)
	GetProviderList(ctx context.Context) ([]Provider, error)

	Search(ctx context.Context) (<-chan SearchResult, error)
}

// Provider structure used by the API
type Provider struct {
	Name string // Provider handle name
}

type SearchResult struct {
	Num   int
	Title string
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
