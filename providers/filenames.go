package providers

import "strings"

var (
	fileNameReplacer = strings.NewReplacer("/", "-", "\\", "-", "!", "", "?", "", ":", "-", ",", "", "*", "-", "|", "-", "\"", "", ">", "", "<", "")
	pathNameReplacer = strings.NewReplacer("!", "", "?", "", ":", "-", ",", "", "*", "", "|", "-", "\"", "", ">", "", "<", "")
)

func FileNameCleaner(s string) string {
	return strings.TrimSpace(fileNameReplacer.Replace(s))
}

func PathNameCleaner(s string) string {
	if i := strings.Index(s, ":"); i >= 0 && i < 2 {
		return s[:i] + strings.TrimSpace(pathNameReplacer.Replace(s[i:]))
	}
	return strings.TrimSpace(pathNameReplacer.Replace(s))
}

func Format2Digits(d string) string {
	if len(d) < 2 {
		return "0" + d
	}
	return d
}
