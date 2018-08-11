package m3u8

import (
	"bufio"
	"fmt"
	"io"
	"github.com/simulot/aspiratv/net/http"
	"strconv"
	"strings"
)

type Getter interface {
	Get(uri string) (io.Reader, error)
}

type Master struct {
	Variants []Variant
	getter   Getter
	URL      string
}

type Variant struct {
	Bandwidth     int64
	Width, Height int64
	worstURL      int64
	URL           string
}

func NewMaster(URL string, getter Getter) (*Master, error) {
	if getter == nil {
		getter = http.DefaultClient
	}
	m := &Master{
		getter: getter,
		URL:    URL,
	}
	r, err := m.getter.Get(URL)
	if err != nil {
		return nil, err
	}
	err = m.decode(r)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Master) WorstQuality() string {
	worstURL := -1
	worstPix := int64(^uint64(0) >> 1)
	for i, v := range m.Variants {
		if v.worstURL < worstPix {
			worstURL = i
			worstPix = v.worstURL
		}
	}
	return http.Rel(m.URL, m.Variants[worstURL].URL)
}

func (m *Master) BestQuality() string {
	bestURL := -1
	bestPix := int64(0)
	for i, v := range m.Variants {
		if v.worstURL > bestPix {
			bestURL = i
			bestPix = v.worstURL
		}
	}
	return http.Rel(m.URL, m.Variants[bestURL].URL)
}

func (m *Master) decode(r io.Reader) error {
	s := bufio.NewScanner(r)
	var v *Variant
	var err error
	waitURL := false
	for s.Scan() {
		l := s.Text()
		if waitURL {
			v.URL = l
			m.Variants = append(m.Variants, *v)
			waitURL = false
		}
		if strings.HasPrefix(l, "#EXT-X-STREAM-INF:") {
			v, err = handleStreamInf(l)
			if err != nil {
				return err
			}
			waitURL = true
			continue
		}

	}
	if err := s.Err(); err != nil && err != io.EOF {
		return err
	}
	return nil
}

func handleStreamInf(s string) (*Variant, error) {
	v := &Variant{}
	var err error
	s = s[len("#EXT-X-STREAM-INF:"):]
	p := splitParams(s)
	for k, val := range p {
		switch k {
		case "BANDWIDTH":
			v.Bandwidth, err = strconv.ParseInt(val, 0, 64)
			if err != nil {
				return nil, fmt.Errorf("Can't parse BANDWIDTH: %v", err)
			}
		case "RESOLUTION":
			_, err = fmt.Sscanf(val, "%dx%d", &v.Width, &v.Height)
			if err != nil {
				return nil, fmt.Errorf("Can't parse RESOLUTION: %v", err)
			}
			v.worstURL = v.Width * v.Height
		}
	}
	return v, nil
}

func splitParams(s string) map[string]string {
	p := 0
	params := map[string]string{}
	for p < len(s) {
		eq := strings.Index(s[p:], "=")
		k := s[p : p+eq]
		p += eq + 1
		if s[p] == '"' {
			q := strings.Index(s[p+1:], `"`)
			params[k] = s[p+1 : p+q+1]
			p += q + 3
		} else {
			q := strings.Index(s[p:], ",")
			if q == -1 {
				params[k] = s[p:]
				p += len(s[p:])
			} else {
				params[k] = s[p : p+q]
				p += q + 1
			}
		}
	}
	return params
}
