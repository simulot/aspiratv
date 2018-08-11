package http

import (
	"strings"
)

// Base returns the base url of a given url.
// It works with url or path
func Base(url string) string {
	// Search the last path separator of the url
	i := len(url) - 1
	for i >= 0 && (url[i] != '/' && url[i] != '\\') {
		i--
	}
	if i >= 0 {
		return url[:i+1]
	}
	return ""
}

func schema(path string) string {
	l := len(path)
	if l < 6 {
		return ""
	}
	if l > 6 {
		l = 6
	}
	s := strings.ToLower(path[:l])
	switch {
	case strings.HasPrefix(s, "https:"):
		return path[:len("https:")]
	case strings.HasPrefix(s, "http:"):
		return path[:len("http:")]
	}
	return ""
}

// IsAbs return true with the url is absolute path
func IsAbs(url string) bool {
	// Check . path
	if len(url) > 0 && url[0] == '.' {
		return false
	}

	// Check for URL schema
	if schema(url) != "" {
		return true
	}

	// Check for unix path
	if len(url) > 0 && url[0] == '/' {
		return true
	}

	// Check windows path
	i := strings.IndexByte(url, '\\')
	if i >= 0 && i <= 2 {
		return true
	}
	return false
}

// Rel return the path of target relative to base.
func Rel(base, target string) string {
	if !IsAbs(target) {
		target = Base(base) + target
	}
	return target
}
