package jscript

import (
	"regexp"
	"unicode/utf8"
)

// AnchorIndex finds in the bytes the first occurrence of the anchor, and
// gives back the index of the 1st '{' after the beginning of the anchor.
// If the pattern isn't found or if the opening bracket isn't found, the function returns -1.
func AnchorIndex(b []byte, anchor *regexp.Regexp) int {
	i := anchor.FindIndex(b)
	if i == nil {
		return -1
	}
	p := i[0]
	for {
		if b[p] == '{' {
			break
		}
		if p == len(b) {
			return -1
		}
		p++
	}

	return p
}

// FindObjectEnd parses the buffer b from objectStart position to the closing '}' respecting structure nesting and strings.
// It returns - when the closing bracket isn't found in the buffer
func FindObjectEnd(b []byte, objectStart int) int {
	nesting := 0
	inDoubleQuoteString := false
	inSimpleQuoteString := false
	lastWasEscape := false

	if objectStart > len(b) || b[objectStart] != '{' {
		return -1
	}

	p := objectStart + 1

	for {
		r, s := utf8.DecodeRune(b[p:])
		switch r {
		case '\\':
			if !lastWasEscape {
				lastWasEscape = true
			}
		case '"':
			if lastWasEscape || inSimpleQuoteString {
				break
			}

			inDoubleQuoteString = !inDoubleQuoteString
			break
		case '\'':
			if lastWasEscape || inDoubleQuoteString {
				break
			}
			inSimpleQuoteString = !inSimpleQuoteString
			break
		case '}':
			if inSimpleQuoteString || inDoubleQuoteString {
				break
			}
			if nesting == 0 {
				return p + 1
			}
			nesting--
		case '{':
			if inSimpleQuoteString || inDoubleQuoteString {
				break
			}
			nesting++
		}
		if lastWasEscape && r != '\\' {
			lastWasEscape = false
		}
		p += s
		if p >= len(b) {
			break
		}
	}
	return -1
}

// ObjectAtAnchor searches an object starting at given anchor and
// returns a slice of bytes containing the object.
// It returns nil when the object is not found, or when the object is not correctly defined.
func ObjectAtAnchor(b []byte, anchor *regexp.Regexp) []byte {
	start := AnchorIndex(b, anchor)
	if start < 0 {
		return nil
	}
	b = b[start:]
	end := FindObjectEnd(b, 0)
	if end < 0 {
		return nil
	}
	return b[:end]
}
