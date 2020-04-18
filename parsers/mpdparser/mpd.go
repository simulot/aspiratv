package mpdparser

import "encoding/xml"

type MPD struct {
	XMLName xml.Name `xml:"urn:mpeg:dash:schema:mpd:2011 MPD"`
	// Xmlns   string   `xml:"xmlns,attr"`
	// XsiName                   string   `xml:"http://www.w3.org/2001/XMLSchema-instance name,attr"`
	SchemaLocation string `xml:"http://www.w3.org/2001/XMLSchema-instance schemaLocation,attr"`
	// Cenc                      string   `xml:"cenc,attr"`
	// Mas                       string   `xml:"xmlns:mas,attr"`
	Type                      string    `xml:"type,attr"`
	MediaPresentationDuration string    `xml:"mediaPresentationDuration,attr"`
	MaxSegmentDuration        string    `xml:"maxSegmentDuration,attr"`
	MinBufferTime             string    `xml:"minBufferTime,attr"`
	Profiles                  string    `xml:"profiles,attr"`
	Period                    []*Period `xml:"Period"`
}

type Period struct {
	ID            string           `xml:"id,attr,omitempty"`
	Duration      string           `xml:"duration,attr,omitempty"`
	BaseURL       string           `xml:"BaseURL,omitempty"`
	AdaptationSet []*AdaptationSet `xml:"AdaptationSet"`
}

type AdaptationSet struct {
	ID                        string                     `xml:"id,attr"`
	Group                     string                     `xml:"group,attr,omitempty"`
	ContentType               string                     `xml:"contentType,attr,omitempty"`
	Lang                      string                     `xml:"lang,attr,omitempty,omitempty"`
	SegmentAlignment          string                     `xml:"segmentAlignment,attr,omitempty"`
	AudioSamplingRate         string                     `xml:"audioSamplingRate,attr,omitempty"`
	MimeType                  string                     `xml:"mimeType,attr,omitempty"`
	Codecs                    string                     `xml:"codecs,attr,omitempty"`
	StartWithSAP              string                     `xml:"startWithSAP,attr,omitempty"`
	Par                       string                     `xml:"par,attr,omitempty"`
	MinBandwidth              string                     `xml:"minBandwidth,attr,omitempty"`
	MaxBandwidth              string                     `xml:"maxBandwidth,attr,omitempty"`
	MaxWidth                  string                     `xml:"maxWidth,attr,omitempty"`
	MaxHeight                 string                     `xml:"maxHeight,attr,omitempty"`
	Sar                       string                     `xml:"sar,attr,omitempty"`
	FrameRate                 string                     `xml:"frameRate,attr,omitempty"`
	AudioChannelConfiguration *AudioChannelConfiguration `xml:"AudioChannelConfiguration,omitempty"`
	Role                      []Role                     `xml:"Role,omitempty"`
	SegmentTemplate           SegmentTemplate            `xml:"SegmentTemplate,omitempty"`
	Representation            []*Representation          `xml:"Representation"`
}

type AudioChannelConfiguration struct {
	SchemeIdUri string `xml:"schemeIdUri,attr"`
	Value       string `xml:"value,attr"`
}

type Role struct {
	SchemeIdUri string `xml:"schemeIdUri,attr"`
	Value       string `xml:"value,attr"`
}

type SegmentTemplate struct {
	Timescale       string `xml:"timescale,attr"`
	Initialization  string `xml:"initialization,attr"`
	Media           string `xml:"media,attr"`
	SegmentTimeline struct {
		S []struct {
			N string `xml:"t,attr,omitempty"`
			T string `xml:"t,attr,omitempty"`
			D string `xml:"d,attr,omitempty"`
			R string `xml:"r,attr,omitempty"`
		} `xml:"S"`
	} `xml:" ,omitempty"`
}

type Representation struct {
	ID        string `xml:"id,attr"`
	Bandwidth string `xml:"bandwidth,attr,omitempty"`
	Width     string `xml:"width,attr,omitempty"`
	Height    string `xml:"height,attr,omitempty"`
	Codecs    string `xml:"codecs,attr,omitempty"`
	ScanType  string `xml:"scanType,attr,omitempty"`
}
