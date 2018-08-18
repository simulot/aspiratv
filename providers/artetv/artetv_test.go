package artetv

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/simulot/aspiratv/net/http/httptest"
)

func readPlayer() (*player, error) {
	f, err := os.Open(filepath.Join("testdata", "player.json"))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	d := json.NewDecoder(f)
	player := &player{}
	err = d.Decode(player)
	if err != nil {
		return nil, err
	}

	return player, nil
}
func TestBestStream(t *testing.T) {
	p, err := New()
	if err != nil {
		t.Error(err)
		return
	}

	player, err := readPlayer()
	if err != nil {
		t.Error(err)
		return
	}

	s := p.getBestVideo(player.VideoJSONPlayer.VSR)
	if s != player.VideoJSONPlayer.VSR["HTTPS_SQ_1"].URL {
		t.Errorf("Unexpected value, got %v", player.VideoJSONPlayer.VSR[s])
	}
}

func Test_sortMapStrInt(t *testing.T) {
	type args struct {
		m mapStrInt
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"empty",
			args{
				mapStrInt{},
			},
			[]string{},
		},
		{
			"one",
			args{
				mapStrInt{"one": 1},
			},
			[]string{"one"},
		},
		{
			"two",
			args{
				mapStrInt{"one": 1, "two": 2},
			},
			[]string{"two", "one"},
		},
		{
			"three",
			args{
				mapStrInt{"one": 1, "two": 2, "three": 3},
			},
			[]string{"three", "two", "one"},
		},
		{
			"three-2",
			args{
				mapStrInt{"two": 2, "three": 3, "one": 1},
			},
			[]string{"three", "two", "one"},
		},
		{
			"three-3",
			args{
				mapStrInt{"three": 3, "one": 1, "two": 2},
			},
			[]string{"three", "two", "one"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sortMapStrInt(tt.args.m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sortMapStrInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sliceIndex(t *testing.T) {
	type args struct {
		k  string
		ls []string
	}
	tests := []struct {
		name string
		args args
		want uint64
	}{
		{
			"empty-key-list",
			args{"", []string{}},
			0,
		},
		{
			"empty-key",
			args{"", []string{"foo", "bar"}},
			0,
		},
		{
			"empty-list",
			args{"foo", []string{}},
			0,
		},
		{
			"test-notfound",
			args{"notfound", []string{"one", "two", "three"}},
			0,
		},
		{
			"test-one",
			args{"one", []string{"one", "two", "three"}},
			1,
		},
		{
			"test-two",
			args{"two", []string{"one", "two", "three"}},
			2,
		},
		{
			"test-three",
			args{"three", []string{"one", "two", "three"}},
			3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sliceIndex(tt.args.k, tt.args.ls); got != tt.want {
				t.Errorf("sliceIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_reverseSliceIndex(t *testing.T) {
	type args struct {
		k  string
		ls []string
	}
	tests := []struct {
		name string
		args args
		want uint64
	}{
		{
			"empty-key-list",
			args{"", []string{}},
			0,
		},
		{
			"empty-key",
			args{"", []string{"foo", "bar"}},
			0,
		},
		{
			"empty-list",
			args{"foo", []string{}},
			0,
		},
		{
			"test-notfound",
			args{"notfound", []string{"one", "two", "three"}},
			0,
		},
		{
			"test-one",
			args{"one", []string{"one", "two", "three"}},
			3,
		},
		{
			"test-two",
			args{"two", []string{"one", "two", "three"}},
			2,
		},
		{
			"test-three",
			args{"three", []string{"one", "two", "three"}},
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := reverseSliceIndex(tt.args.k, tt.args.ls); got != tt.want {
				t.Errorf("reverseSliceIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getGuide(t *testing.T) {
	getter := httptest.New(httptest.WithConstantFile(filepath.Join("testdata", "guide.json")))

	p, err := New(withGetter(getter))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	ss, err := p.getGuide(nil, time.Time{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if len(ss) != 31 {
		t.Errorf("Expecting %d shows, got %d", 31, len(ss))
	}
}

func Test_getPlayerResolutions(t *testing.T) {
	player, err := readPlayer()
	if err != nil {
		t.Error(err)
		return
	}
	res := getPlayerResolutions(player.VideoJSONPlayer.VSR)
	if len(res) != 4 {
		t.Errorf("Expected resolution number %d, got %d", 4, len(res))
	}
}

func TestArteTV_getStreamScore(t *testing.T) {

	player, err := readPlayer()
	if err != nil {
		t.Error(err)
		return
	}

	sortedResolution := getPlayerResolutions(player.VideoJSONPlayer.VSR)

	type args struct {
		s               streamInfo
		resolutionIndex uint64
	}
	tests := []struct {
		name string
		args args
		want uint64
	}{
		{
			"HTTPS_SQ_1",
			args{
				player.VideoJSONPlayer.VSR["HTTPS_SQ_1"],
				reverseSliceIndex(getResolutionKey(player.VideoJSONPlayer.VSR["HTTPS_SQ_1"]), sortedResolution),
			},
			1004010,
		},
		{
			"HLS_XQ_1",
			args{
				player.VideoJSONPlayer.VSR["HLS_XQ_1"],
				reverseSliceIndex(getResolutionKey(player.VideoJSONPlayer.VSR["HLS_XQ_1"]), sortedResolution),
			},
			1004000,
		},
		{
			"HTTPS_HQ_1",
			args{
				player.VideoJSONPlayer.VSR["HTTPS_HQ_1"],
				reverseSliceIndex(getResolutionKey(player.VideoJSONPlayer.VSR["HTTPS_HQ_1"]), sortedResolution),
			},
			1002010,
		},
		{
			"HTTPS_SQ_2",
			args{
				player.VideoJSONPlayer.VSR["HTTPS_SQ_2"],
				reverseSliceIndex(getResolutionKey(player.VideoJSONPlayer.VSR["HTTPS_SQ_2"]), sortedResolution),
			},
			4010,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ArteTV{
				preferredVersions: []string{"VF", "VOF", "VOSTF", "VF-STF"},
				preferredMedia:    "mp4",
			}
			if got := p.getStreamScore(tt.args.s, tt.args.resolutionIndex); got != tt.want {
				t.Errorf("ArteTV.getStreamScore() = %v, want %v", got, tt.want)
			}
		})
	}
}
