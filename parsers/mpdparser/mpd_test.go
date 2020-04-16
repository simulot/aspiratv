package mpdparser

import (
	"encoding/xml"
	"io/ioutil"
	"path"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMPDReadWrite(t *testing.T) {
	tests, err := ioutil.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		if path.Ext(tt.Name()) == ".mpd" {
			t.Run(tt.Name(), func(t *testing.T) {
				original, err := ioutil.ReadFile(path.Join("testdata", tt.Name()))
				if err != nil {
					t.Fatal(err)
				}
				mpdRead := &MPD{}
				err = xml.Unmarshal(original, mpdRead)
				if err != nil {
					t.Error(err)
					return
				}

				output, err := xml.MarshalIndent(mpdRead, "", "  ")
				if err != nil {
					t.Error(err)
					return
				}
				output = append([]byte(xml.Header), output...)
				ioutil.WriteFile(path.Join("testdata", tt.Name()+".out"), output, 0777)

				mpdGenerated := &MPD{}
				err = xml.Unmarshal(output, mpdGenerated)
				if err != nil {
					t.Error(err)
					return
				}

				// dmp := diffmatchpatch.New()
				// diffs := dmp.DiffMain(string(original), string(output), false)
				// t.Log(string(output), "\n", dmp.DiffPrettyText(diffs))
				if diff := cmp.Diff(mpdRead, mpdGenerated); diff != "" {
					t.Errorf("MDP read write read test mismatch (-want +got):\n%s", diff)
				}
			})
		}
	}
}
