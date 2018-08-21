package artetv

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/alecthomas/repr"
)

func mustParse(t *testing.T, f string, d string) time.Time {
	date, err := time.Parse(f, d)
	if err != nil {
		t.Fatal(err)
	}
	return date
}
func Test_readDetails(t *testing.T) {
	expected := &showInfo{
		airDate:  mustParse(t, "2006-01-02", "2018-07-02"),
		season:   "2017",
		title:    "La minute vieille",
		subTitle: "Pulsion irréstistible",
	}

	f, err := os.Open(filepath.Join("testdata", "minute.html.tst"))
	if err != nil {
		t.Fatal(err)
		return
	}
	defer f.Close()

	o, err := readDetails(f)
	if err != nil {
		t.Error(err)
		return
	}
	if !reflect.DeepEqual(expected, o) {
		t.Errorf("Expected %s, got %s", repr.String(expected), repr.String(o))
	}
}
