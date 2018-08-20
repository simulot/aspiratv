package jscript

import (
	"io/ioutil"
	"path"
	"regexp"
	"testing"
)

func TestPlayerjs(t *testing.T) {
	f, err := ioutil.ReadFile(path.Join("testdata", "player.js"))
	if err != nil {
		t.Error(err)
		return
	}
	o, err := ParseObjectAtAnchor(f, regexp.MustCompile(`exports=\{[a-z]{2}:\{Club`))
	if err != nil {
		t.Error(err)
		return
	}

	s := o.Property("fr").Property("Club").Property("profile").String()
	if s != "Modifier mon profil" {
		t.Errorf("Expected s to be %q, got %q", "Modifier mon profil", s)

	}

}
