package providers

import (
	"context"

	"github.com/simulot/aspiratv/models"
)

type Provider interface {
	Search(ctx context.Context, q models.SearchQuery) (<-chan models.SearchResult, error)
	ProviderDescribe(ctx context.Context) ProviderDescription
}

type ProviderDescription struct {
	Name     string
	Channels []string
}
