package nfo

import (
	"strings"
)

var (
	fileNameReplacer = strings.NewReplacer("/", "-", "\\", "-", "!", "", "?", "", ":", "", "*", "-", "|", "-", "\"", "", ">", "", "<", "")
	pathNameReplacer = strings.NewReplacer("!", "", "?", "", ":", " ", ",", "", "*", "", "|", " ", "\"", "", ">", "", "<", "")
)

// FileNameCleaner return a safe file name from a given show name.
func FileNameCleaner(s string) string {
	return strings.TrimSpace(fileNameReplacer.Replace(s))
}

// PathNameCleaner return a safe path name from a given show name.
func PathNameCleaner(s string) string {
	if i := strings.Index(s, ":"); i >= 0 && i < 2 {
		return s[:i] + strings.TrimSpace(pathNameReplacer.Replace(s[i:]))
	}
	return strings.TrimSpace(pathNameReplacer.Replace(s))
}

// Format2Digits return a number with 2 digits when there is only one digit
func Format2Digits(d string) string {
	if len(d) < 2 {
		return "0" + d
	}
	return d
}
