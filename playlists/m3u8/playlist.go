package m3u8

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/simulot/aspiratv/net/http"
)

type Playlist struct {
	Duration   time.Duration
	URL        string
	base       string
	allowCache bool
	chunks     []chunk
	getter     Getter
}

type chunk struct {
	duration time.Duration
	url      string
}

func NewPlayList(url string, getter Getter) (*Playlist, error) {
	if getter == nil {
		getter = http.DefaultClient
	}
	p := &Playlist{
		URL:    url,
		chunks: []chunk{},
		getter: getter,
	}

	r, err := p.getter.Get(url)
	if err != nil {
		return nil, err
	}
	err = p.decode(r)
	return p, err
}

func (p *Playlist) decode(r io.Reader) error {
	s := bufio.NewScanner(r)
	var c *chunk
	waitURL := false
	for s.Scan() {
		l := s.Text()
		if strings.HasPrefix(l, "#EXT-X-ALLOW-CACHE:") {
			v := l[len("#EXT-X-ALLOW-CACHE:"):]
			p.allowCache = v == "YES"
			continue
		}
		if strings.HasPrefix(l, "#EXTINF:") {
			d := float64(0)
			_, err := fmt.Sscanf(l, "#EXTINF:%f", &d)
			if err != nil {
				return fmt.Errorf("Can't parse chunk of playlist: %v", err)
			}
			c = &chunk{
				duration: time.Duration(d * float64(time.Second)),
			}
			waitURL = true
			continue
		}
		if waitURL {
			var err error
			c.url = l
			if err != nil {
				return fmt.Errorf("Can't parse chunk of playlist: %v", err)
			}
			p.chunks = append(p.chunks, *c)
			p.Duration += c.duration
			waitURL = false
		}
	}
	if s.Err() != io.EOF {
		return s.Err()
	}
	return nil
}

func (p *Playlist) Download() (io.Reader, error) {
	pr, pw := io.Pipe()
	go func() {
		for _, c := range p.chunks {
			url := c.url
			if !http.IsAbs(c.url) {
				url = http.Base(p.URL) + c.url
			}
			r, err := p.getter.Get(url)
			if err != nil {
				pw.CloseWithError(err)
				return
			}
			_, err = io.Copy(pw, r)
			if err != nil {
				pw.CloseWithError(err)
				return
			}
		}
		pw.Close()
	}()
	return pr, nil

}
