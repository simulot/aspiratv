package query

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func Test_Structure(t *testing.T) {
	files := []string{
		// "cretins.json",
		"season.json",
	}

	for _, f := range files {
		b, err := ioutil.ReadFile(filepath.Join("testdata", f))
		if err != nil {
			t.Fatal(err)
			return
		}

		r := QueryResults{}
		err = json.Unmarshal(b, &r)
		if err != nil {
			t.Fatal(err)
			return
		}
		t.Logf("%#v", r)
	}

}
