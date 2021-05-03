package providers

import (
	"context"

	"github.com/simulot/aspiratv/pkg/models"
)

type Provider interface {
	Name() string
	Search(ctx context.Context, q models.SearchQuery) (<-chan models.SearchResult, error)
	ProviderDescribe(ctx context.Context) Description
	GetMedias(ctx context.Context, task models.DownloadTask) (<-chan models.DownloadItem, error)
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
