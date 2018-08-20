package jscript

import (
	"fmt"
	"regexp"
)

// ParseObjectAtAnchor locate the object assignent at the anchor and return a Structure with its content
func ParseObjectAtAnchor(b []byte, anchor *regexp.Regexp) (*Structure, error) {
	b = ObjectAtAnchor(b, anchor)
	if b == nil {
		return nil, fmt.Errorf("Can't find object in the buffer")
	}
	o, err := ParseObject(b)
	if err != nil {
		return nil, err
	}
	return o, nil

}

// Property get object property's value
func (s *Structure) Property(name string) *Value {
	for _, p := range s.Properties {
		if p.Name == name {
			return p.Value
		}
	}
	return nil
}

// String return string when the Value has type String
func (v *Value) String() string {
	if v.Str != nil {
		return *v.Str
	}
	return ""
}

// Null return true when the Value has type Null and is null
func (v *Value) Null() bool {
	if v.NullStr != nil {
		return *v.NullStr == "null"
	}
	return false
}

// Property return the value of the property named "name"
func (v *Value) Property(name string) *Value {
	if v.Struct != nil {
		return v.Struct.Property(name)
	}
	return nil
}

// Strings return a slice of string when the value has type Array of string
func (v *Value) Strings() []string {
	a := []string{}
	for _, s := range v.Ar {
		if s.Str != nil {
			a = append(a, *s.Str)
		}
	}
	return a
}
