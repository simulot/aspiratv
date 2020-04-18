package download

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/simulot/aspiratv/metadata/nfo"
)

type dummyProgresser struct {
	size  int64
	count int64
}

func (p *dummyProgresser) Init(size int64) {
	p.size = size
}

func (p *dummyProgresser) Update(count int64, size int64) {
	p.count = count
	p.size = size
	fmt.Printf("%.1f%% %d / %d   \n", float64(count)/float64(p.size)*100.0, count, size)
}

func TestFFMepg(t *testing.T) {
	type args struct {
		ctx           context.Context
		in            string
		out           string
		info          *nfo.MediaInfo
		configurators []ffmpegConfigurator
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			"Normal",
			args{
				ctx:  context.Background(),
				in:   "https://file-examples.com/wp-content/uploads/2017/04/file_example_MP4_640_3MG.mp4",
				out:  os.DevNull,
				info: &nfo.MediaInfo{},
				configurators: []ffmpegConfigurator{
					FFMpegWithDebug(true),
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := FFMpeg(tt.args.ctx, tt.args.in, tt.args.out, tt.args.info, tt.args.configurators...); (err != nil) != tt.wantErr {
				t.Errorf("FFMepg() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
