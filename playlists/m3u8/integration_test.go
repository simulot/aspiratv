// +build integration

package m3u8

import (
	"io"
	"os/exec"
	"testing"
)

func TestDownload(t *testing.T) {
	testCases := []struct {
		name   string
		md5sum string
	}{
		{"https://bitdash-a.akamaihd.net/content/MI201109210084_1/m3u8s/f08e80da-bf1d-4e3d-8899-f0f6155f6efa.m3u8", "3f5ce2b21252a21722ae17f26a90164e"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// getter := nil // &fileGet{}

			m, err := NewMaster(tc.name, nil)
			if err != nil {
				t.Fatal(err)
				return
			}

			p, err := NewPlayList(m.WorstQuality(), nil)
			if err != nil {
				t.Fatal(err)
				return
			}
			r, err := p.Download()
			if err != nil {
				t.Fatal(err)
				return
			}

			cmd := exec.Command("mplayer", "-cache", "1024", "-")
			stdin, err := cmd.StdinPipe()
			if err != nil {
				t.Fatal(err)
				return
			}
			go func() {
				if _, err := io.Copy(stdin, r); err != nil {
					t.Fatalf("Can't pipe to comand: %v", err)
					stdin.Close()
					return
				}
				stdin.Close()
			}()
			err = cmd.Run()
			if err != nil {
				t.Errorf("Can't run command: %v", err)
				return
			}

		})

	}

}
