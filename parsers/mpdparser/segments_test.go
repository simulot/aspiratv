package mpdparser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_segmentPostion_Format(t *testing.T) {
	type fields struct {
		RepresentationID string
		Number           int
		Time             int
		// Duration         int
		// StartNumber      int
		// TimeScale        int
	}
	f := fields{
		"video_50000",
		1000,
		123456,
	}

	tests := []struct {
		name     string
		template string
		want     string
	}{
		// TODO: Add test cases.
		{
			"empty string",
			"",
			"",
		},
		{
			"no variable",
			"26991696_monde_TA-RepresentationID-Time.dash",
			"26991696_monde_TA-RepresentationID-Time.dash",
		},
		{
			"variable RepresentationID",
			"26991696_monde_TA-$RepresentationID$-Time.dash",
			"26991696_monde_TA-video_50000-Time.dash",
		},
		{
			"variable time",
			"26991696_monde_TA-RepresentationID-$Time$.dash",
			"26991696_monde_TA-RepresentationID-123456.dash",
		},
		{
			"variable time at beginning",
			"$Time$_26991696_monde_TA-RepresentationID.dash",
			"123456_26991696_monde_TA-RepresentationID.dash",
		},
		{
			"variable time at end",
			"26991696_monde_TA-RepresentationID.dash_$Time$",
			"26991696_monde_TA-RepresentationID.dash_123456",
		},
		{
			"variables RepresentationID time and number ",
			"226991696_monde_TA-$RepresentationID$-$Time$.$Number$.dash",
			"226991696_monde_TA-video_50000-123456.1000.dash",
		},
		{
			"unknown variable",
			"26991696_monde_TA-$Unknown$-Time.dash",
			"26991696_monde_TA-$Unknown$-Time.dash",
		},
		{
			"Unpaired $",
			"26991696_monde_TA-$Time-Time.dash",
			"26991696_monde_TA-$Time-Time.dash",
		},
		{
			"Unpaired $ at begining",
			"$26991696_monde_TA-Time-Time.dash",
			"$26991696_monde_TA-Time-Time.dash",
		},
		{
			"Unpaired $ at end",
			"26991696_monde_TA-Time-Time.dash$",
			"26991696_monde_TA-Time-Time.dash$",
		},
		{
			"Unpaired $ at begining and a variable",
			"$26991696_monde_TA-$Time$.dash",
			"$26991696_monde_TA-123456.dash",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := segmentPostion{
				RepresentationID: f.RepresentationID,
				Number:           f.Number,
				Time:             f.Time,
				// Duration:         tt.fields.Duration,
				// StartNumber:      tt.fields.StartNumber,
				// TimeScale:        tt.fields.TimeScale,
			}

			if got := s.Format(tt.template); got != tt.want {
				t.Errorf("segmentPostion.Format() = %v, want %v", got, tt.want)
			}
		})
	}
}

// $ cat show4.log | sed -rne "s/.*Opening '.*\/(.*)' for reading/\"\1\",/gp" | grep video >show4_video.log

func parseLog(name string, filters []string) (map[string][]string, error) {
	var re = regexp.MustCompile(`(?m).*Opening '(.*)' for reading`)
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := map[string][]string{}
	for _, v := range filters {
		r[v] = []string{}
	}

	s := bufio.NewScanner(f)
	for s.Scan() {
		m := re.FindStringSubmatch(s.Text())
		if len(m) >= 2 {
			for _, f := range filters {
				if strings.Contains(m[0], f) {
					r[f] = append(r[f], string(m[1]))
				}
			}
		}
	}
	err = s.Err()
	if err != nil {
		return nil, err
	}
	return r, nil
}

func mustParseLog(t *testing.T, name string, filters []string) map[string][]string {
	r, err := parseLog(name, filters)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func Test_SegmentList(t *testing.T) {
	type test struct {
		name                    string
		log                     string
		URLBase                 string
		audioAdaptationSetMime  string
		audioRepresentationID   string
		videoAdapatationSetMime string
		videoRepresentationID   string
		readLimit               int
	}
	tests := []test{
		{
			name:                    "testdata/show4.mpd",
			log:                     "testdata/show4.log",
			URLBase:                 "https://cloudreplayfrancetv.akamaized.net/0da1d04f889e5/226991696_monde_TA.ism/ZXhwPTE1ODcyNDc2MTR+YWNsPSUyZjBkYTFkMDRmODg5ZTUlMmYyMjY5OTE2OTZfbW9uZGVfVEEuaXNtKn5obWFjPWQxNWU5MTEwOGY5MDBlZWJlYWIxNjE3ZmE3YjE4NDM2OTU4YWUyZTJlNDA2MmUwNWQyYzBiMWVhYmU2NjIyZmM=/manifest.mpd",
			audioAdaptationSetMime:  "audio/mp4",
			audioRepresentationID:   "audio_fre=64000",
			videoAdapatationSetMime: "video/mp4",
			videoRepresentationID:   "video=400000",
			readLimit:               0,
		},
		{
			name:                    "testdata/show3.mpd",
			log:                     "testdata/show3.log",
			URLBase:                 "https://bitmovin-a.akamaihd.net/content/MI201109210084_1/mpds/f08e80da-bf1d-4e3d-8899-f0f6155f6efa.mpd",
			audioAdaptationSetMime:  "audio/mp4",
			audioRepresentationID:   "1_stereo_128000",
			videoAdapatationSetMime: "video/mp4",
			videoRepresentationID:   "180_250000",
			readLimit:               50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// read log file
			segments := mustParseLog(t, tt.log, []string{"audio", "video"})
			mpd := NewMPDParser()

			f, err := os.Open(tt.name)
			if err != nil {
				t.Fatal(err)
			}
			err = mpd.Read(f)
			if err != nil {
				t.Fatal(err)
			}

			as := mpd.Period[0].GetAdaptationSetByMimeType(tt.audioAdaptationSetMime)
			if as == nil {
				t.Error("Adaptation set not found")
				return
			}
			r := as.GetRepresentationByID(tt.audioRepresentationID)
			if r == nil {
				t.Error("Representation  not found")
				return
			}

			it, err := mpd.MediaURIs(tt.URLBase, mpd.Period[0], as, r)
			if err != nil {
				t.Fatal(err)
			}
			got, err := pullSegments(it, tt.readLimit)
			if err != nil {
				t.Fatal(err)
			}

			want := segments["audio"]
			if tt.readLimit > 0 {
				want = want[:tt.readLimit]
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Segments generated and expected mismatch (-want +got):\n%s", diff)
			}

			as = mpd.Period[0].GetAdaptationSetByMimeType(tt.videoAdapatationSetMime)
			if as == nil {
				t.Error("Adaptation set not found")
				return
			}
			r = as.GetRepresentationByID(tt.videoRepresentationID)
			if r == nil {
				t.Error("Representation  not found")
				return
			}
			it, err = mpd.MediaURIs(tt.URLBase, mpd.Period[0], as, r)
			if err != nil {
				t.Fatal(err)
			}
			got, err = pullSegments(it, tt.readLimit)
			if err != nil {
				t.Fatal(err)
			}

			want = segments["video"]
			if tt.readLimit > 0 {
				want = want[:tt.readLimit]
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Segments generated and expected mismatch (-want +got):\n%s", diff)
			}

			// if diff, mesg := compareSlices(got, want); diff {
			// 	t.Error(mesg)
			// }

			_ = segments
		})
	}

}

func pullSegments(it SegmentIterator, limit int) ([]string, error) {
	r := []string{}
	i := 0
	for s := range it.Next() {
		if limit != 0 && i >= limit {
			it.Cancel()
			break
		}

		if s.Err != nil {
			return nil, s.Err
		}
		r = append(r, s.S)
		i++
	}

	return r, nil
}

func compareSlices(got, want []string) (bool, string) {
	if len(want) != len(got) {
		return true, fmt.Sprintf("Got len is %d, expected %d", len(got), len(want))
	}

	for i := 0; i < len(got); i++ {
		if want[i] != got[i] {
			return true, fmt.Sprintf("Index %d:\n got[%d]=%q\nwant[%d]=%q", i, i, got[i], i, want[i])
		}
	}
	return false, ""
}
