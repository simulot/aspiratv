package main

import (
	"os"
	"path/filepath"
	"testing"
)

func getwd() string {
	r, _ := os.Getwd()
	return r
}

func Test_SanitizePath(t *testing.T) {
	os.Setenv("HOME", "/home/user")
	cc := []struct {
		name string
		arg  string
		want string
		good bool
	}{
		{
			"simple",
			os.ExpandEnv("$HOME/Video/DL"),
			filepath.Join("/home/user", "Video/DL"),
			true,
		},
		{
			"relative to WD",
			os.ExpandEnv("./DL"),
			filepath.Join(getwd(), "DL"),
			true,
		},
		{
			"parent of wd",
			os.ExpandEnv("../DL"),
			"",
			false,
		},
		{
			"parent",
			os.ExpandEnv("../.././../../../../../../../../../etc/passwd"),
			"",
			false,
		},
		{
			"parent and environement",
			os.ExpandEnv("$HOME/../../../../../../../../../etc/passwd"),
			"",
			false,
		},
	}

	for _, c := range cc {
		t.Run(c.name, func(t *testing.T) {
			got, err := sanitizePath(c.arg)
			if (err == nil) != c.good {
				t.Errorf("Expecting error:%v, got error %s", c.good, err)
			}
			if got != c.want {
				t.Errorf("Expecting return:%q, got %q", c.want, got)
			}
		})
	}
}
