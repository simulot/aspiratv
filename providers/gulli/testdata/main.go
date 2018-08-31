package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
)

// var reVars = regexp.MustCompile(
// 	`(?P<sources>sources:)` +
// 		`|(?:file:\s*(?U:"(?P<file>[^"]*)"))` +
// 		`|(?:mediaid:\s*(?U:"(?P<mediaid>[^"]*)"))` +
// 		`|(?:playlist_title:\s*(?U:"(?P<playlist_title>[^"]*)"))` +
// 		`|(?:image:\s*(?U:"(?P<image>[^"]*)"))`)

var reVars = regexp.MustCompile(
	`(?P<sources>sources:)` +
		`|(?:file:\s*(?U:"(?P<file>.*)"))` +
		`|(?:mediaid:\s*(?U:"(?P<mediaid>.*)"))` +
		`|(?:playlist_title:\s*(?U:"(?P<playlist_title>.*)"))` +
		`|(?:image:\s*(?U:"(?P<image>.*)"))`)

func main() {
	f, err := os.Open("player_VOD68993752052000.html.txt")
	if err != nil {
		os.Exit(1)
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		os.Exit(1)
	}
	parts := reVars.SubexpNames()
	s := string(b)
	m := reVars.FindAllStringSubmatch(s, -1)
	fmt.Println(len(m))
	for i, sm := range m {
		for j, s := range sm {
			if j > 0 && len(s) > 0 {
				fmt.Println(i, parts[j], s)
			}
		}
	}

}
