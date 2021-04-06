package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"sync"

	"github.com/simulot/aspiratv/parsers/mpdparser"
)

func main() {

	// trap Ctrl+C and call cancel on the context
	ctx, cancel := context.WithCancel(context.Background())
	breakChannel := make(chan os.Signal, 1)
	signal.Notify(breakChannel, os.Interrupt)

	defer func() {
		// Normal end... cleaning up
		signal.Stop(breakChannel)
		cancel()
	}()

	// waiting for interruption
	go func() {
		select {
		case <-breakChannel:
			cancel()
		case <-ctx.Done():
			return
		}
	}()

	destination := flag.String("out", "", "output name")
	flag.Parse()

	if flag.NArg() == 0 {
		log.Println("Missing manifest url")
		flag.Usage()
	}
	if len(*destination) == 0 {
		log.Println("Missing output name")
		flag.Usage()

	}

	manifest := flag.Arg(0)

	mpd := mpdparser.NewMPDParser()
	err := mpd.Get(ctx, manifest)
	if err != nil {
		log.Printf("Can't get manifest: %s", err)
		os.Exit(1)
	}

	audioAS := mpd.Period[0].GetAdaptationSetByMimeType("audio/mp4")
	bestAudio := audioAS.GetBestRepresentation()
	audioIT, err := mpd.MediaURIs(manifest, mpd.Period[0], audioAS, bestAudio)

	if err != nil {
		log.Printf("Can't get audio segments list: %s", err)
		os.Exit(1)
	}

	videoAS := mpd.Period[0].GetAdaptationSetByMimeType("video/mp4")
	bestVideo := videoAS.GetBestRepresentation()
	videoIT, err := mpd.MediaURIs(manifest, mpd.Period[0], videoAS, bestVideo)

	if err != nil {
		log.Printf("Can't get video segments list: %s", err)
		os.Exit(1)
	}

	// subtitleAS := mpd.Period[0].GetAdaptationSetByMimeType("application/mp4")
	// bestSubtitle := subtitleAS.Representation[0]
	// subtitleIT, err := mpd.MediaURIs(manifest, mpd.Period[0], subtitleAS, bestSubtitle)
	if err != nil {
		log.Printf("Can't get video segments list: %s", err)
		os.Exit(1)
	}

	concurency := make(chan bool, 3)
	concurency <- true
	concurency <- true
	ctx2, cancel2 := context.WithCancel(ctx)
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		err := download(ctx2, *destination+".video.mp4", videoIT, concurency)
		if err != nil {
			log.Println(err)
			cancel2()
		}
		wg.Done()
	}()
	go func() {
		err := download(ctx2, *destination+".audio.mp4", audioIT, concurency)
		if err != nil {
			log.Println(err)
			cancel2()
		}
		wg.Done()
	}()
	// go func() {
	// 	err := download(ctx2, *destination+".subtitles.sub", subtitleIT, concurency)
	// 	if err != nil {
	// 		log.Println(err)
	// 		cancel2()
	// 	}
	// 	wg.Done()
	// }()
	wg.Wait()
	if err := ctx2.Err(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	cancel2()

	cmd := exec.Command("ffmpeg",
		"-i", *destination+".video.mp4",
		"-i", *destination+".audio.mp4",
		"-codec", "copy",
		"-f", "mp4",
		"-y",
		*destination,
	)
	err = cmd.Run()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	os.Remove(*destination + ".audio.mp4")
	os.Remove(*destination + ".video.mp4")
	log.Print("done.")

}

func download(ctx context.Context, filename string, it mpdparser.SegmentIterator, concurency chan bool) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

loop:
	for s := range it.Next() {
		<-concurency
		select {
		case <-ctx.Done():
			break loop
		default:
			if s.Err != nil {
				fmt.Fprintf(os.Stderr, "Can't range video segment: %s", s.Err)
				os.Exit(1)
			}
			log.Println("Get ", s.S)
			r, err := http.Get(s.S)
			if err != nil {
				return err
			}
			if r.StatusCode >= 200 && r.StatusCode < 300 {
				_, err = io.CopyBuffer(f, r.Body, nil)
			}
			if r.StatusCode >= 400 {
				log.Printf("Error %s when getting %s", r.Status, s.S)
			}
			r.Body.Close()
			if r.StatusCode == 404 {
				it.Cancel()
				break loop
			}
			concurency <- true
		}
	}

	return nil

}
