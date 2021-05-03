package francetv

import (
	"encoding/json"
	"io"
	"os"
	"testing"
)

func TestSearch(t *testing.T) {
	decoded := loadFile(t, "data/marleau.json")

	var found map[string]Result
	err := json.Unmarshal([]byte(decoded), &found)
	if err != nil {
		t.Fatalf("Can't decode the structure: %s", err)
		return
	}

	if len(found) == 0 {
		t.Fatal("No result")
	}

	tx := found["taxonomy"]

	if len(tx.Hits) == 0 {
		t.Fatal("No Hits")
	}
	hit := tx.Hits[0]

	if hit.Class == "" {
		missing(t, "hit.Class missing")
	}
	if hit.URLComplete == "" {
		missing(t, "hit.URLComplete missing")
	}

	if hit.Channel == "" {
		missing(t, "hit.Channel missing")
	}

	if len(hit.Image.Formats) == 0 {
		missing(t, "hit.Image.Formats missing")
	}

	if hit.Image.Formats["vignette_16x9"].Urls["w:1024"] == "" {
		missing(t, "hit.Image.Formats[vignette_16x9].Url[w:1024] missing")
	}
}

func TestTitle(t *testing.T) {
	decoded := loadFile(t, "data/aviation.json")
	var found map[string]Result
	err := json.Unmarshal([]byte(decoded), &found)
	if err != nil {
		t.Fatalf("Can't decode the structure: %s", err)
		return
	}

	for _, c := range found {
		for _, h := range c.Hits {
			if h.ID == 2312411 {
				if h.Program.Label == "" {
					missing(t, ".Program.Label")
				}
			}
		}
	}

}

func missing(t *testing.T, field string) {
	t.Helper()
	t.Errorf("Field %s is missing", field)
}

func loadFile(t *testing.T, name string) string {
	t.Helper()
	f, err := os.Open(name)
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		t.Error(err)
	}
	return string(b)
}
