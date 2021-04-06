package main

import (
	"fmt"
	"os"

	"github.com/simulot/aspiratv/parsers/ttml"
)

func dieIfErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
func main() {
	in, err := os.Open(os.Args[1])
	dieIfErr(err)
	out, err := os.Create(os.Args[1] + ".srt")
	dieIfErr(err)
	_, err = ttml.TrancodeToSRT(out, in)
	out.Sync()
	dieIfErr(err)
	in.Close()
	out.Close()
}
