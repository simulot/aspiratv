package m3u8

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"text/template"
)

func TestHandleStreamInf(t *testing.T) {
	testCases := []struct {
		s             string
		bandwidth     int64
		width, height int64
	}{
		{`#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=258157,CODECS="avc1.4d400d,mp4a.40.2",AUDIO="stereo",RESOLUTION=422x180,SUBTITLES="subs"`,
			258157,
			422,
			180,
		},
		{`#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=873000,RESOLUTION=704x396,CODECS="avc1.77.30, mp4a.40.2"`,
			873000,
			704,
			396,
		},
		{`#EXT-X-STREAM-INF:PROGRAM-ID=1,RESOLUTION=704x396,CODECS="avc1.77.30, mp4a.40.2",BANDWIDTH=873000`,
			873000,
			704,
			396,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.s, func(t *testing.T) {
			m, err := handleStreamInf(tc.s)
			if err != nil {
				t.Error(err)
			}
			if m.Bandwidth != tc.bandwidth {
				t.Errorf("Expected bandwidth to be %d, go %d", tc.bandwidth, m.Bandwidth)
			}
			if m.Width != tc.width {
				t.Errorf("Expected width to be %d, go %d", tc.width, m.Width)
			}
			if m.Height != tc.height {
				t.Errorf("Expected height to be %d, go %d", tc.height, m.Height)
			}
		})
	}

}

func TestMalformedHandleStreamInf(t *testing.T) {
	testCases := []struct {
		s             string
		bandwidth     int64
		width, height int64
	}{
		{`#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=258157,CODECS="avc1.4d400d,mp4a.40.2",AUDIO="stereo",RESOLUTION=422180,SUBTITLES="subs"`,
			258157,
			422,
			180,
		},
		{`#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=@zazjé,RESOLUTION=704x396,CODECS="avc1.77.30, mp4a.40.2"`,
			873000,
			704,
			396,
		},
		{`#EXT-X-STREAM-INF:PROGRAM-ID=1,RESOLUTION=704X396,CODECS="avc1.77.30, mp4a.40.2",BANDWIDTH=873000`,
			873000,
			704,
			396,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.s, func(t *testing.T) {
			_, err := handleStreamInf(tc.s)
			if err != nil {
				t.Log(err)
			}
			if err == nil {
				t.Errorf("Expecting an error, but recieve nil")
			}
		})
	}

}

func TestMasterWorstBestQuality(t *testing.T) {

	testCases := []struct {
		name  string
		best  string
		worst string
	}{
		{
			"testdata/f2-rg/master.m3u8",
			"testdata/f2-rg/index_3_av.m3u8",
			"testdata/f2-rg/index_0_av.m3u8",
		},
		{
			"testdata/akamai/master.m3u8",
			"testdata/akamai/10000kbit.m3u8",
			"testdata/akamai/250kbit.m3u8",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			getter := &fileGet{}
			ctx := context.TODO()

			m, err := NewMaster(ctx, tc.name, getter)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			best := m.BestQuality()
			if tc.best != best {
				t.Errorf("Expected best url to be '%s', got '%s'", tc.best, best)
			}
			worst := m.WorstQuality()
			if tc.worst != worst {
				t.Errorf("Expected worst url to be '%s', got '%s'", tc.worst, worst)
			}
		})
	}

}

type fileGet struct {
}

func (*fileGet) Get(ctx context.Context, url string) (io.ReadCloser, error) {
	f, err := os.Open(url)
	if err != nil {
		return nil, err
	}

	pr, pw := io.Pipe()
	go func() {
		defer f.Close()
		_, err := io.Copy(pw, f)
		if err != nil {
			pw.CloseWithError(err)
			return
		}
		pw.Close()
	}()

	return pr, nil
}

type stringGet struct {
	s                map[int]string
	md5              string
	playlist, master strings.Builder
	expected         strings.Builder
}

func newStringGet(n int) *stringGet {
	s := &stringGet{
		s: make(map[int]string),
	}

	h := md5.New()

	for i := 0; i < n; i++ {
		c := 'A' + i%26
		s.s[i] = strings.Repeat(string(c), 50)
		s.expected.WriteString(s.s[i])
		h.Write([]byte(s.s[i]))
	}
	s.md5 = fmt.Sprintf("%x", h.Sum(nil))

	tmpl, err := template.New("master.m3u8").Parse(
		`#EXTM3U
#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=258157,CODECS="avc1.4d400d,mp4a.40.2",AUDIO="stereo",RESOLUTION=422x180,SUBTITLES="subs"
playlist.m3u8

`)
	if err != nil {
		panic("Can't parse master.m3u8 template")
	}
	err = tmpl.Execute(&s.master, nil)
	if err != nil {
		panic("Can't Execute master.m3u8 template")
	}
	tmpl, err = template.New("playlist.m3u8").Parse(
		`#EXTM3U
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-TARGETDURATION:{{len .}}
{{- range $i,$s := .}}
#EXTINF:2,
seq-{{$i }}.ts
{{- end}}

#EXT-X-ENDLIST

`)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(&s.playlist, s.s)
	if err != nil {
		panic("Can't Execute playlist.m3u8 template")
	}
	// fmt.Println(s.master.String())
	// fmt.Println(s.playlist.String())
	return s
}

func (s *stringGet) Get(ctx context.Context, url string) (io.ReadCloser, error) {
	switch {
	case url == "master.m3u8":
		return ioutil.NopCloser(strings.NewReader(s.master.String())), nil
	case url == "playlist.m3u8":
		return ioutil.NopCloser(strings.NewReader(s.playlist.String())), nil
	}
	var i int
	_, err := fmt.Sscanf(url, "seq-%d.ts", &i)
	if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(strings.NewReader(s.s[i])), nil
}

func TestMultiPartPlayList(t *testing.T) {
	testCases := []int{
		1,
		2,
		50,
		500,
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Test with %d chunks", tc), func(t *testing.T) {
			getter := newStringGet(tc)
			ctx := context.TODO()

			m, err := NewMaster(ctx, "master.m3u8", getter)
			if err != nil {
				t.Fatal(err)
				return
			}
			p, err := NewPlayList(ctx, m.BestQuality(), getter)
			if err != nil {
				t.Fatal(err)
				return
			}

			r, err := p.Download(ctx)
			if err != nil {
				t.Fatal(err)
				return
			}

			result := &strings.Builder{}
			if _, err := io.Copy(result, r); err != nil {
				t.Fatalf("Can't get data from stream: %v", err)
				return
			}
			if result.String() != getter.expected.String() {
				t.Errorf("Expecting content: %s, got: %s\n", getter.expected.String(), result.String())
			}

		})
	}
}
