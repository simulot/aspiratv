package download

import "context"

type Downloader interface {
	Download(ctx context.Context, destination string) (err error)
	WithProgresser(p Progresser)
	WithLogger(l interface{ Printf(string, ...interface{}) })
}
