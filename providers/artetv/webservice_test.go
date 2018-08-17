package artetv

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// Test decoding of guide structure
func TestGuide(t *testing.T) {
	f, err := os.Open(filepath.Join("testdata", "guide.json"))
	if err != nil {
		t.Error(err)
		return
	}

	d := json.NewDecoder(f)
	guide := &guide{}
	err = d.Decode(guide)
	if err != nil {
		t.Error(err)
		return
	}

}

// Test player decoding
func TestPlayer(t *testing.T) {
	f, err := os.Open(filepath.Join("testdata", "player.json"))
	if err != nil {
		t.Error(err)
		return
	}

	d := json.NewDecoder(f)
	player := &player{}
	err = d.Decode(player)
	if err != nil {
		t.Error(err)
		return
	}
}
