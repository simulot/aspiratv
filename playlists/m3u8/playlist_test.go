package m3u8

import (
	"testing"
	"time"
)

func TestPlayListContent(t *testing.T) {
	testCases := []struct {
		name          string
		totalDuration time.Duration
		count         int
		allowCache    bool
	}{
		{"testdata/f2-rg/index_3_av.m3u8",
			time.Duration(time.Second*16726 + 613*time.Millisecond),
			1673,
			true,
		},
		{"testdata/akamai/250kbit.m3u8",
			time.Duration(time.Second * 444 * 2),
			444,
			false,
		}}

	for _, tc := range testCases {
		getter := &fileGet{}

		p, err := NewPlayList(tc.name, getter)
		if err != nil {
			t.Fatal(err)
			return
		}
		if tc.count != len(p.chunks) {
			t.Errorf("Expecting chunk count to be %d, but got %d", tc.count, len(p.chunks))
		}
		if tc.totalDuration != p.Duration {
			t.Errorf("Expecting total duration to be %s, but got %s", tc.totalDuration, p.Duration)
		}
		if tc.allowCache != p.allowCache {
			t.Errorf("Expecting allowCache to be %v, but got %v", tc.allowCache, p.allowCache)
		}

	}
}
