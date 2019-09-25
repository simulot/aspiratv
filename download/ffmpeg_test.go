package download

import (
	"context"
	"testing"
)

func TestFFMepg(t *testing.T) {
	type args struct {
		u   string
		prg Progresser
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"hls",
			args{
				"http://qthttp.apple.com.edgesuite.net/1010qwoeiuryfg/sl.m3u8",
				&dummyProgresser{},
			},
			false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := FFMepg(context.Background(), tt.args.u, tt.args.prg); (err != nil) != tt.wantErr {
				t.Errorf("FFMepg() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
