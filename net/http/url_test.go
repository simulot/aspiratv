package http

import "testing"

func TestBase(t *testing.T) {
	testCases := []struct {
		g, e string
	}{
		{"/hello/world.txt", "/hello/"},
		{"/hello/", "/hello/"},
		{"/hello", "/"},
		{"/", "/"},
		{"hello.txt", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.g, func(t *testing.T) {
			r := Base(tc.g)
			if r != tc.e {
				t.Errorf("Expecting %q, got %q", tc.e, r)
			}
		})
	}
}

func TestIsAbs(t *testing.T) {
	testCases := []struct {
		g string
		e bool
	}{
		{"https://hello/world.txt", true},
		{"http://hello/world.txt", true},
		{"/hello/world.txt", true},
		{"/hello/world.txt", true},
		{"/hello/world.txt", true},
		{"hello/world.txt", false},
		{"\\hello\\world.txt", true},
		{"hello\\world.txt", false},
		{"C:\\hello\\world.txt", true},
		{"./hello/world.txt", false},
		{"../hello/world.txt", false},
		{"/world.txt", true},
		{"\\world.txt", true},
		{"world.txt", false},
		{"./world.txt", false},
		{".\\world.txt", false},
		{"./a.c", false},
	}
	for _, tc := range testCases {
		t.Run(tc.g, func(t *testing.T) {
			r := IsAbs(tc.g)
			if r != tc.e {
				t.Errorf("Expecting %v, got %v", tc.e, r)
			}
		})
	}
}

func TestRel(t *testing.T) {
	testCases := []struct {
		base, target, expected string
	}{
		{"", "hello.txt", "hello.txt"},
		{"path/", "hello.txt", "path/hello.txt"},
		{"", "/hello.txt", "/hello.txt"},
		{"path/", "/hello.txt", "/hello.txt"},
		{"http://path/", "/hello.txt", "/hello.txt"},
	}

	for _, tc := range testCases {
		t.Run(tc.base, func(t *testing.T) {
			result := Rel(tc.base, tc.target)
			if result != tc.expected {
				t.Errorf("Expecting %q, got %q", tc.expected, result)
			}
		})
	}
}
