package nfo

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
)

type Season struct {
	XMLName      xml.Name `xml:"season"`
	Plot         string   `xml:"plot,omitempty"`
	Outline      string   `xml:"outline,omitempty"`
	Lockdata     string   `xml:"lockdata,omitempty"`
	Dateadded    string   `xml:"dateadded,omitempty"`
	Title        string   `xml:"title,omitempty"`
	Year         string   `xml:"year,omitempty"`
	Aired        Aired    `xml:"premiered,omitempty"`
	Seasonnumber string   `xml:"seasonnumber,omitempty"`
	Thumb        []Thumb  `xml:"-"`
}

func (n *Season) WriteNFO(destination string) error {
	err := os.MkdirAll(filepath.Dir(destination), 0777)
	if err != nil {
		return fmt.Errorf("Can't create %s :%w", destination, err)
	}

	f, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("Can't create tvshow.nfo :%w", err)
	}
	defer f.Close()

	_, err = f.WriteString(xml.Header)
	err = xml.NewEncoder(f).Encode(n)
	if err != nil {
		return fmt.Errorf("Can't encode tvshow.nfo :%w", err)
	}

	return nil
}
