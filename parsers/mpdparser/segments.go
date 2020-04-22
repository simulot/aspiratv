package mpdparser

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/purell"
	//"github.com/PuerkitoBio/purell"
)

func (mpd *MPDParser) getAdaptationBaseURL(manifestURL string, p *Period, a *AdaptationSet) (string, error) {
	if len(manifestURL) > 0 {
		if manifestURL[len(manifestURL)-1] != '/' {
			i := strings.LastIndex(manifestURL, "/")
			if i > 0 {
				manifestURL = manifestURL[:i]
			}
		}
	}
	return purell.NormalizeURLString(manifestURL+"/"+p.BaseURL, purell.FlagsUsuallySafeNonGreedy)

}

func (mpd *MPDParser) getSegmentDuration(p *Period, s *SegmentTemplate) (int, error) {
	if s.Duration > 0 {
		return s.Duration, nil
	}

	d, err := mpd.getPresentationDuration(p)
	if err != nil {
		return 0, err
	}
	return int(d.Seconds() * float64(s.Timescale)), nil
}

func (mpd *MPDParser) getPresentationDuration(p *Period) (time.Duration, error) {
	s := ""
	if len(p.Duration) > 0 {
		s = p.Duration
	} else {
		s = mpd.MPD.MediaPresentationDuration
	}
	return GetPTasDuration(s)
}

// GetPTasDuration gets presentation time as time.Duration
func GetPTasDuration(pt string) (time.Duration, error) {
	if len(pt) == 0 {
		return time.Duration(0), errors.New("Period time duration empty")
	}
	re := regexp.MustCompile(`(?m)^P(\d+Y)?(\d+M)?(\d+D)?T(\d+H)?(\d+M)?([0-9.]+S)?$`)
	d := time.Duration(0)
	for _, match := range re.FindAllStringSubmatch(pt, -1) {
		for _, m := range match {

			if len(m) < 2 {
				continue
			}
			if m[0] == 'P' {
				continue
			}

			s := m[:len(m)-1]
			u := m[len(m)-1]
			switch u {
			// Ignores Y,M,D duration
			case 'H':
				v, err := strconv.Atoi(s)
				if err != nil {
					return time.Duration(0), fmt.Errorf("Invalid period duration:%s", pt)
				}
				d += time.Duration(v) * time.Hour
			case 'M':
				v, err := strconv.Atoi(s)
				if err != nil {
					return time.Duration(0), fmt.Errorf("Invalid period duration:%s", pt)
				}
				d += time.Duration(v) * time.Minute
			case 'S':
				v, err := strconv.ParseFloat(s, 64)
				if err != nil {
					return time.Duration(0), fmt.Errorf("Invalid period duration:%s", pt)
				}
				d += time.Duration(v * float64(time.Second))
			}
		}
	}
	return d, nil
}

type SegmentIterator interface {
	Next() <-chan SegmentItem // Call it until error == io.EOF
	Cancel()                  // To stop the iterator before the end
	Err() error
}

func (mpd *MPDParser) MediaURIs(ManifestURL string, p *Period, a *AdaptationSet, r *Representation) (SegmentIterator, error) {
	if a.SegmentTemplate != nil && len(a.SegmentTemplate.SegmentTimeline.S) > 0 {
		return mpd.mediaByTimeLine(ManifestURL, p, a, r)
	}
	return mpd.mediaByNumber(ManifestURL, p, a, r)
}

type segmentPostion struct {
	RepresentationID string // $RepresentationID$
	Number           int    // $Number$ = Current segment number
	Time             int    // $Time$ = current time in presentation units (not seconds)
	Duration         int
	StartNumber      int
	TimeScale        int
}

func (s segmentPostion) Format(template string) string {
	var b strings.Builder
	p := 0
	for p < len(template) {
		// Before variable
		i := strings.Index(template[p:], "$")
		if i < 0 {
			// No more variable
			b.WriteString(template[p:])
			break
		}
		if i > 0 {
			b.WriteString(template[p : p+i])
		}

		if p+i+1 > len(template) {
			b.WriteByte(template[p+i])
			break
		}
		p = p + i + 1
		i = strings.Index(template[p:], "$")
		if i < 0 {
			//  variable not terminated
			b.WriteString(template[p-1:])
			break
		}
		switch strings.ToLower(template[p : p+i]) {
		case "representationid":
			b.WriteString(s.RepresentationID)
		case "number":
			b.WriteString(strconv.Itoa(s.Number))
		case "time":
			b.WriteString(strconv.Itoa(s.Time))
		default:
			b.WriteByte('$')
			b.WriteString(template[p : p+i])
			i--
		}
		p = p + i + 1
	}
	return b.String()
}

func normalizeSegmentURL(base, segment string) (string, error) {
	return purell.NormalizeURLString(base+segment, purell.FlagsUsuallySafeGreedy)
}

type SegmentItem struct {
	S        string
	Position segmentPostion
	Err      error
}

type timeLineIterator struct {
	mpd         *MPDParser
	manifestURL string
	p           *Period
	a           *AdaptationSet
	r           *Representation
	pos         segmentPostion
	err         error
	segmentChan chan SegmentItem
	closeChan   chan interface{}
}

func (i *timeLineIterator) Next() <-chan SegmentItem {
	return i.segmentChan
}

func (i *timeLineIterator) send(u string, position segmentPostion, err error) bool {
	select {
	case <-i.closeChan:
		return false
	case i.segmentChan <- SegmentItem{u, position, err}:
		return true
	}
}

func (i *timeLineIterator) Cancel() {
	select {
	case i.closeChan <- struct{}{}:

	default:
		close(i.segmentChan)
	}
}

func (i timeLineIterator) Err() error {
	return i.err
}

func (mpd *MPDParser) mediaByTimeLine(manifestURL string, p *Period, a *AdaptationSet, r *Representation) (*timeLineIterator, error) {

	base, err := mpd.getAdaptationBaseURL(manifestURL, p, a)
	if err != nil {
		return nil, err
	}

	d, err := mpd.getSegmentDuration(p, a.SegmentTemplate)
	if err != nil {
		return nil, err
	}

	it := &timeLineIterator{
		mpd:         mpd,
		manifestURL: manifestURL,
		p:           p,
		a:           a,
		r:           r,
		pos: segmentPostion{
			RepresentationID: r.ID,
			TimeScale:        a.SegmentTemplate.Timescale,
			Number:           a.SegmentTemplate.StartNumber,
			Duration:         d,
		},
		segmentChan: make(chan SegmentItem),
		closeChan:   make(chan interface{}),
	}

	go func() {
		defer it.Cancel()

		// The init
		u, err := normalizeSegmentURL(base, it.pos.Format(it.a.SegmentTemplate.Initialization))
		if ok := it.send(u, it.pos, err); !ok || err != nil {
			return
		}

		// Segments
		for _, s := range a.SegmentTemplate.SegmentTimeline.S {
			if it.pos.Time > it.pos.Duration {
				return
			}
			for i := 0; i < s.R+1; i++ {
				if i == 0 && s.T > 0 {
					it.pos.Time = s.T
				}
				u, err := normalizeSegmentURL(base, it.pos.Format(a.SegmentTemplate.Media))
				if ok := it.send(u, it.pos, err); !ok || err != nil {
					return
				}

				it.pos.Time += s.D
				it.pos.Number++
			}
		}
	}()
	return it, nil
}

type numberIterator struct {
	mpd         *MPDParser
	manifestURL string
	p           *Period
	a           *AdaptationSet
	r           *Representation
	pos         segmentPostion
	err         error
	segmentChan chan SegmentItem
	closeChan   chan interface{}
}

func (it *numberIterator) Next() <-chan SegmentItem {
	return it.segmentChan
}

func (it *numberIterator) send(u string, position segmentPostion, err error) bool {
	select {
	case <-it.closeChan:
		return false
	case it.segmentChan <- SegmentItem{u, position, err}:
		return true
	}
}

func (it *numberIterator) Cancel() {
	select {
	case <-it.closeChan:
	default:
		close(it.closeChan)
	}
}

func (it numberIterator) Err() error {
	return it.err
}

func (mpd *MPDParser) mediaByNumber(manifestURL string, p *Period, a *AdaptationSet, r *Representation) (*numberIterator, error) {

	base, err := mpd.getAdaptationBaseURL(manifestURL, p, a)
	if err != nil {
		return nil, err
	}

	d, err := mpd.getSegmentDuration(p, a.SegmentTemplate)
	if err != nil {
		return nil, err
	}

	it := &numberIterator{
		mpd:         mpd,
		manifestURL: manifestURL,
		p:           p,
		a:           a,
		r:           r,
		pos: segmentPostion{
			RepresentationID: r.ID,
			TimeScale:        a.SegmentTemplate.Timescale,
			Number:           a.SegmentTemplate.StartNumber,
			Duration:         d,
		},
		segmentChan: make(chan SegmentItem),
		closeChan:   make(chan interface{}),
	}

	go func() {
		defer it.Cancel()
		// The init
		u, err := normalizeSegmentURL(base, it.pos.Format(it.a.SegmentTemplate.Initialization))
		if ok := it.send(u, it.pos, err); !ok || err != nil {
			return
		}

		// Segments
		for {
			u, err := normalizeSegmentURL(base, it.pos.Format(a.SegmentTemplate.Media))
			if ok := it.send(u, it.pos, err); !ok || err != nil {
				break
			}

			it.pos.Number++
		}
	}()
	return it, nil
}
