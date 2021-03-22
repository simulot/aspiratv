package ttml

import (
	"encoding/xml"
	"strconv"
	"strings"
)

/*
	This package reads TTML subtitles as delivered by franctev.
	Each chunk begins with some binary data that we will discard and
	regular TTML xml chunk

*/

type TTML struct {
	XMLName xml.Name `xml:"tt"`
	Lang    string   `xml:"lang,attr"`
	// Regions []Region `xml:"head>layout>region"`
	// Styles  []Style  `xml:"head>styling>style"`
	Pages []Page `xml:"body>div>p"`
}

// type Subtitle struct {
// 	ID     string   `xml:"id,attr,omitempty"`
// 	Begin  TTMLTime `xml:"begin,attr,omitempty"`
// 	End    TTMLTime `xml:"end,attr,omitempty"`
// 	Items  string   `xml:",innerxml"` // We must store inner XML here since there's no tag to describe both any tag and chardata
// 	Region string   `xml:"region,attr,omitempty"`
// 	Style  string   `xml:"style,attr,omitempty"`
// 	StyleAttributes
// }

type StyleAttributes struct {
	// BackgroundColor string `xml:"backgroundColor,attr"`
	// DisplayAlign    string `xml:"displayAlign,attr"`
	// Extent          string `xml:"extent,attr"`
	// FontFamily      string `xml:"fontFamily,attr"`
	// FontSize        string `xml:"fontSize,attr"`
	// Origin          string `xml:"origin,attr"`
	// TextAlign       string `xml:"textAlign,attr"`
	Color string `xml:"color,attr,omitempty"`

	// TextOutline     string `xml:"textOutline,attr"`
	// BackgroundColor string `xml:"backgroundColor,attr,omitempty"`
	// Direction       string `xml:"direction,attr,omitempty"`
	// Display         string `xml:"display,attr,omitempty"`
	// DisplayAlign    string `xml:"displayAlign,attr,omitempty"`
	// Extent          string `xml:"extent,attr,omitempty"`
	// FontFamily      string `xml:"fontFamily,attr,omitempty"`
	// FontSize        string `xml:"fontSize,attr,omitempty"`
	// FontStyle       string `xml:"fontStyle,attr,omitempty"`
	// FontWeight      string `xml:"fontWeight,attr,omitempty"`
	// LineHeight      string `xml:"lineHeight,attr,omitempty"`
	// Opacity         string `xml:"opacity,attr,omitempty"`
	// Origin          string `xml:"origin,attr,omitempty"`
	// Overflow        string `xml:"overflow,attr,omitempty"`
	// Padding         string `xml:"padding,attr,omitempty"`
	// ShowBackground  string `xml:"showBackground,attr,omitempty"`
	// TextAlign       string `xml:"textAlign,attr,omitempty"`
	// TextDecoration  string `xml:"textDecoration,attr,omitempty"`
	// TextOutline     string `xml:"textOutline,attr,omitempty"`
	// UnicodeBidi     string `xml:"unicodeBidi,attr,omitempty"`
	// Visibility      string `xml:"visibility,attr,omitempty"`
	// WrapOption      string `xml:"wrapOption,attr,omitempty"`
	// WritingMode     string `xml:"writingMode,attr,omitempty"`
	// ZIndex          int    `xml:"zIndex,attr,omitempty"`
}

// type Style struct {
// 	ID string `xml:"id,attr"`
// 	StyleAttributes
// }

// type Region struct {
// 	ID    string `xml:"id,attr"`
// 	Style string `xml:"style,attr"`
// }

type Page struct {
	Begin TTMLTime `xml:"begin,attr"`
	End   TTMLTime `xml:"end,attr"`
	// Region string   `xml:"region,attr"`
	ID   Caption `xml:"id,attr"`
	Span []struct {
		Text string `xml:",chardata"`
		StyleAttributes
	} `xml:"span"`
}

type TTMLTime string

func (t *TTMLTime) UnmarshalText(i []byte) error {
	*t = TTMLTime(strings.ReplaceAll(string(i), ".", ","))
	return nil
}

type Caption int

func (c *Caption) UnmarshalText(b []byte) error {
	i, err := strconv.Atoi(strings.TrimPrefix(string(b), "caption"))
	*c = Caption(i)
	return err
}
