package mpdparser

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/purell"
)

// MPDParser holds the MPD structure
type MPDParser struct {
	*MPD
	ActualURL string
}

// NewMPDParser allocate a new MPD parser
func NewMPDParser() *MPDParser {
	return &MPDParser{}
}

// Get queries the webserver, remember the redirected location and get the MPD
func (p *MPDParser) Get(ctx context.Context, url string) error {
	// Set up the HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// disable redirection system using the simple vanilla roundtripper
	transport := http.Transport{}
	resp, err := transport.RoundTrip(req)
	if err != nil {
		return err
	}
	// Check if you received the status codes you expect.
	if resp.StatusCode >= 400 {
		return fmt.Errorf("Failed with status: %q", resp.Status)
	}

	if resp.StatusCode >= 300 {
		// Get the new location, and jump on it
		p.ActualURL = resp.Header.Get("Location")
		resp.Body.Close()

		resp, err = http.Get(p.ActualURL)
		if err != nil {
			return err
		}
		if resp.StatusCode >= 400 {
			return fmt.Errorf("Failed with status: %q", resp.Status)
		}
	}

	return p.Read(resp.Body)
}

// Read the MPD from the reader and close it
func (p *MPDParser) Read(rc io.ReadCloser) error {
	b, err := ioutil.ReadAll(rc)
	if err != nil {
		return err
	}
	defer rc.Close()
	return p.Unmarshal(b)
}

// Unmarshal bytes as MPD
func (p *MPDParser) Unmarshal(b []byte) error {
	mpd := &MPD{}
	err := xml.Unmarshal(b, mpd)
	if err != nil {
		return fmt.Errorf("Can't unmarshal MPD: %w", err)
	}

	p.MPD = mpd
	return nil
}

func (p *MPDParser) Marshal() ([]byte, error) {
	output, err := xml.MarshalIndent(p.MPD, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("Can't marshal MPD: %w", err)
	}
	return output, nil
}

// Write the MPD to the writer and close it
func (p *MPDParser) Write(wc io.WriteCloser) error {
	defer wc.Close()
	output, err := p.Marshal()
	if err != nil {
		return fmt.Errorf("Can't marshal MPD: %w", err)
	}
	_, err = wc.Write([]byte(xml.Header))
	if err != nil {
		return fmt.Errorf("Can't marshal MPD: %w", err)
	}
	_, err = wc.Write(output)
	if err != nil {
		return fmt.Errorf("Can't write MPD: %w", err)
	}
	return nil
}

// StripSTPPStream isn't correctly handeld by FFMPEG at writing time
func (p *MPDParser) StripSTPPStream() error {
	for _, pe := range p.MPD.Period {
		newAS := []*AdaptationSet{}
		for _, a := range pe.AdaptationSet {
			if a.MimeType == "application/mp4" && a.Codecs == "stpp" {
				continue
			}
			newAS = append(newAS, a)
		}
		pe.AdaptationSet = newAS
	}
	return nil
}

// AbsolutizeURLs change relative urls to absolute
// Either there is a BaseURL, and we have to make it absolute. In this case, segment templates should be relative
// Either there isn't BasURL. Then if Segments templates are relative to MDP's URI
// Check BaseURL or SegmentTemplates
func (p *MPDParser) AbsolutizeURLs(base string) error {
	for _, p := range p.MPD.Period {
		if len(p.BaseURL) > 0 {

			u, err := url.Parse(p.BaseURL)
			if err != nil {
				return fmt.Errorf("Can't parse MPD BaseURL: %w", err)
			}
			if !u.IsAbs() {
				p.BaseURL, err = changeURL(base, p.BaseURL)
				if err != nil {
					return err
				}
			}
			continue
		}
		for _, a := range p.AdaptationSet {
			if len(a.SegmentTemplate.Initialization) > 0 {
				u, err := changeURL(base, a.SegmentTemplate.Initialization)
				if err != nil {
					return fmt.Errorf("Can't process url change: %w", err)
				}
				a.SegmentTemplate.Initialization = u
			}
			if len(a.SegmentTemplate.Media) > 0 {
				u, err := changeURL(base, a.SegmentTemplate.Media)
				if err != nil {
					return fmt.Errorf("Can't process url change: %w", err)
				}
				a.SegmentTemplate.Media = u
			}
		}

	}
	return nil
}

func changeURL(base, URL string) (string, error) {
	u, err := url.Parse(URL)
	if err != nil {
		return "", err
	}
	if u.IsAbs() {
		return URL, nil
	}
	return purell.NormalizeURLString(base+"/"+URL, purell.FlagsUsuallySafeGreedy)
}

// KeepBestVideoStream discard all video stream but the one with the highest bandwidth
func (p *MPDParser) KeepBestVideoStream() error {
	for _, p := range p.MPD.Period {
		for _, a := range p.AdaptationSet {
			if !strings.HasPrefix(a.MimeType, "video/") {
				continue
			}

			bandwith := 0
			best := -1
			for i, r := range a.Representation {
				if r.Bandwidth > bandwith {
					bandwith = r.Bandwidth
					best = i
				}
			}

			if best == -1 {
				return nil
			}
			a.Representation = []*Representation{
				a.Representation[best],
			}
			a.MinBandwidth = a.Representation[0].Bandwidth
			a.MaxBandwidth = a.Representation[0].Bandwidth
			a.MaxWidth = a.Representation[0].Width
			a.MaxHeight = a.Representation[0].Height
		}
	}
	return nil
}
