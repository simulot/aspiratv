package providers

import (
	"context"

	"github.com/simulot/aspiratv/models"
)

type Provider interface {
	Search(ctx context.Context, q models.SearchQuery) (<-chan models.SearchResult, error)
	ProviderDescribe(ctx context.Context) Description
}

type Description struct {
	Code     string
	Name     string
	Logo     string
	URL      string
	Channels map[string]Channel
}

type Channel struct {
	Code string
	Name string
	Logo string
	URL  string
}
