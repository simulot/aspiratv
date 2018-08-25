package jsonparser

import (
	"fmt"
	"strconv"
	"time"
)

// Seconds read a number of seconds and transform it into time.Duration
type Seconds time.Duration

// UnmarshalJSON converts a string number of secondes into a time.Duration
func (s *Seconds) UnmarshalJSON(b []byte) error {
	if b[0] == '"' {
		b = b[1 : len(b)-1]
	}
	if string(b) == "null" || string(b) == "" {
		*s = 0
		return nil
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return fmt.Errorf("Can't parse duration in seconds: %v", err)
	}
	*s = Seconds(time.Duration(i) * time.Second)
	return nil
}

func (s Seconds) Duration() time.Duration { return time.Duration(s) }

// TSUnix read a unix timestamp and transform it into a time.Time
type TSUnix time.Time

// UnmarshalJSON read a unix timestamp and transform it into a time.Time
func (t *TSUnix) UnmarshalJSON(b []byte) error {
	if b[0] == '"' {
		b = b[1 : len(b)-1]
	}
	i, err := strconv.ParseInt(string(b), 0, 64)
	if err != nil {
		return err
	}
	// convert the unix epoch to a Time object
	*t = TSUnix(time.Unix(i, 0))
	return nil
}
