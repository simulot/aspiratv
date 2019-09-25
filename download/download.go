package download

import "context"

type DownloadFn func(ctx context.Context, u string, params []string, prg Progresser) error

func DownloadFunction(probe *FFProbeOutput) DownloadFn {
	return FFMepg
}
