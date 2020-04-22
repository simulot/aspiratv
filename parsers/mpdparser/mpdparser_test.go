package mpdparser

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestFRANCETVMpd(t *testing.T) {
	type adaptation struct {
		mimeType               string
		bandWith               int
		templateInitialization string
		templateMedia          string
	}
	tests := []struct {
		name           string
		videoBandWidth int
		givenBaseURL   string
		baseURL        string
		adaptations    []adaptation
	}{
		// {
		// 	name:           "testdata/mdpsegment.mpd",
		// 	videoBandWidth: "",
		// },
		// {
		// 	name:           "testdata/show1.mpd",
		// 	videoBandWidth: "2000000",
		// 	givenBaseURL:   "https://videoserver.com/2535f38e-a6ae-49b2-a047-c01d7bd03c6e",
		// 	baseURL:        "https://videoserver.com/2535f38e-a6ae-49b2-a047-c01d7bd03c6e/dash/",
		// 	adaptations: []adaptation{
		// 		{
		// 			ID:                     "3",
		// 			bandWith:               "2000000",
		// 			templateInitialization: "226948192_france-domtom_TA-$RepresentationID$.dash",
		// 			templateMedia:          "226948192_france-domtom_TA-$RepresentationID$-$Time$.dash",
		// 		},
		// 	},
		// },
		// {
		// 	name:           "testdata/show2.mpd",
		// 	videoBandWidth: "2000000",
		// },
		{
			name:           "testdata/show3.mpd",
			videoBandWidth: 4800000,
			givenBaseURL:   "https://bitmovin-a.akamaihd.net/content/MI201109210084_1/mpds",
			baseURL:        "",
			adaptations: []adaptation{
				{
					mimeType:               "video/mp4",
					bandWith:               4800000,
					templateInitialization: "https://bitmovin-a.akamaihd.net/content/MI201109210084_1/video/$RepresentationID$/dash/init.mp4",
					templateMedia:          "https://bitmovin-a.akamaihd.net/content/MI201109210084_1/video/$RepresentationID$/dash/segment_$Number$.m4s",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(tt.name)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			tempFile, err := ioutil.TempFile(os.TempDir(), "*_manifest.mpd")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tempFile.Name())

			mpd := NewMPDParser()
			err = mpd.Read(f)
			if err != nil {
				t.Fatal(err)
			}
			err = mpd.StripSTPPStream()
			if err != nil {
				t.Error(err)
			}
			// Check absence of stpp encoded subtitles
			nSTPP := 0
			for _, p := range mpd.Period {
				for _, a := range p.AdaptationSet {
					if a.Codecs == "stpp" {
						nSTPP++
					}
				}
			}
			if nSTPP > 0 {
				t.Errorf("STPP segment still present")
			}

			err = mpd.KeepBestVideoStream()
			if err != nil {
				t.Error(err)
			}
			for _, p := range mpd.Period {
				bandWith := 0
				maxBandWidth := 0
				nVideoStream := 0
				for _, a := range p.AdaptationSet {
					if strings.HasPrefix(a.MimeType, "video/") {
						for _, r := range a.Representation {
							nVideoStream++
							bandWith = r.Bandwidth
						}
						maxBandWidth = a.MaxBandwidth
					}
				}
				if nVideoStream != 1 {
					t.Errorf("Unexpected number of video stream (%d) for Period ID(%s), expected 1", nVideoStream, p.ID)
				}
				if bandWith != tt.videoBandWidth {
					t.Errorf("Unexpected bandwidth (%d) for video stream for Period ID(%s), expected %d", bandWith, p.ID, tt.videoBandWidth)
				}
				if bandWith != maxBandWidth {
					t.Errorf("Unexpected maxBandWidth (%d) for video stream for Period ID(%s), expected %d", maxBandWidth, p.ID, tt.videoBandWidth)
				}

			}

			err = mpd.AbsolutizeURLs(tt.givenBaseURL)
			if err != nil {
				t.Error(err)
			}
			if mpd.MPD.Period[0].BaseURL != tt.baseURL {
				t.Errorf("Unexpected baseURL (%s), expected (%s)", mpd.MPD.Period[0].BaseURL, tt.baseURL)
			}
			for _, a := range tt.adaptations {
				for _, as := range mpd.Period[0].AdaptationSet {
					if as.MimeType != a.mimeType {
						continue
					}

					if a.templateInitialization != as.SegmentTemplate.Initialization {
						t.Errorf("Unexpected SegmentTemplate.Initialization (%s) of representation(AdaptationID:%s), expected (%s)", as.SegmentTemplate.Initialization, as.MimeType, a.templateInitialization)
					}
					if a.templateMedia != as.SegmentTemplate.Media {
						t.Errorf("Unexpected SegmentTemplate.Media (%s) of representation(AdaptationID:%s), expected (%s)", as.SegmentTemplate.Media, as.MimeType, a.templateMedia)
					}
				}
			}

			err = mpd.Write(tempFile)
			if err != nil {
				t.Error(err)
			}
		})
	}
}
