package models

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestMessageJSON(t *testing.T) {
	want := NewMessage("Hello").SetPinned(true).SetStatus(StatusError)
	want.When = want.When.Truncate(0) // Strip the monotonic clock to allow time comparison
	b, err := json.MarshalIndent(want, "", "  ")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	got := &Message{}
	err = json.Unmarshal(b, got)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	got.When = got.When.Truncate(0)
	if !reflect.DeepEqual(want, got) {
		t.Errorf("Expecting \n%#v,\n\tgot\n\t%#v", want, got)
	}
}
