package artetv

import (
	"encoding/json"
	"io"
	"os"
	"testing"
)

func TestExtractState(t *testing.T) {
	f := readFile(t, "data/series-bron.html")
	b, err := extractState(f, "__INITIAL_STATE__")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}

	s := InitialProgram{}
	err = json.Unmarshal(b, &s)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}

}

func readFile(t *testing.T, filename string) []byte {
	t.Helper()
	f, err := os.Open(filename)
	if err != nil {
		t.Error(err)
		return nil
	}

	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		t.Error(err)
		return nil
	}
	return b
}
