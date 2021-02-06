package ttml

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

func TrancodeToSRT(dst io.Writer, src io.Reader) (int64, error) {
	t := srtTrancoder{
		src: bufio.NewReader(src),
		dst: bufio.NewWriter(dst),
	}
	return t.Run()
}

type srtTranscoderState int

const (
	srtNone srtTranscoderState = iota
	srtFirstHeader
	srtHeader
	srtInXMLFragment
	srtDone
)

type srtTrancoder struct {
	state         srtTranscoderState
	dst           *bufio.Writer
	src           *bufio.Reader
	buff          strings.Builder
	captionNumber int
	read          int64
}

type strStateFn func() (strStateFn, error)

func (t *srtTrancoder) Run() (int64, error) {
	defer t.dst.Flush()
	var err error
	fn := t.waittHeader
	for fn != nil {
		fn, err = fn()
		if err == io.EOF {
			break
		}
		if err != nil {
			return t.read, err
		}
	}
	return t.read, nil
}

func (t *srtTrancoder) discardUntilByteSequence(seq []byte) error {
	for {
	restart:
		for _, r := range seq {
			b, err := t.src.ReadByte()
			if err != nil {
				return err
			}
			t.read++
			if r != b {
				goto restart
			}
		}
		// We have reached the seq
		return nil
	}
}

func (t *srtTrancoder) waittHeader() (strStateFn, error) {
	err := t.discardUntilByteSequence([]byte("mdat"))
	if err == io.EOF {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("Can't read TTML, missing  mdat segment: %w", err)
	}

	return t.readXLMFragment, nil
}

func (t *srtTrancoder) readXLMFragment() (strStateFn, error) {
	fragment := []byte{}
	endReached := false
	for !bytes.HasSuffix(fragment, []byte("</tt>")) {
		b, err := t.src.ReadByte()
		if err == io.EOF {
			endReached = true
			break
		}
		if err != nil {
			return nil, err
		}
		t.read++
		fragment = append(fragment, b)
	}

	tt := TTML{}
	err := xml.NewDecoder(bytes.NewReader(fragment)).Decode(&tt)
	if err != nil {
		return nil, fmt.Errorf("Can't parse TTML xml: %w", err)
	}
	err = tt.ToSrt(t.dst)
	if err != nil {
		return nil, fmt.Errorf("Can't convert TTML to Srt: %w", err)
	}

	if endReached {
		err = io.EOF
	}

	return t.waittHeader, err
}

func (tt *TTML) ToSrt(dst io.Writer) error {
	for _, page := range tt.Pages {
		fmt.Fprintln(dst, page.ID)
		fmt.Fprintln(dst, page.Begin, "-->", page.End)
		for _, l := range page.Span {
			if l.Color != "white" {
				fmt.Fprintln(dst, fmt.Sprintf(`<font color="%s">%s</font>`, l.Color, l.Text))
			} else {
				fmt.Fprintln(dst, l.Text)
			}
		}
		fmt.Fprintln(dst)
	}
	return nil
}
