package query

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func Test_Structure(t *testing.T) {
	b, err := ioutil.ReadFile(filepath.Join("testdata", "cretins.json"))
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
