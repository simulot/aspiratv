package jscript

import (
	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
)

// This package helps to read object initialisation from a Javascript source, which is a bit different from JSON.
// Classical JSON parser fails short on unquoted property names.
//
// You have to point the beggining of the object initialisation with a regular expression.
// The routine locates the opening bracket ({) and the closing bracket (}),
// whatever the structure depth.
//
// A generic structure is built with object properties and values.
//
// Implementation details: I'm using Alec Thomas's Participle package that provides a simple parser.
//

// Structure object is a slice of Properties surrounded by brackets
type Structure struct {
	Properties []*Property `"{" [ @@ { "," @@ } ]  "}"`
}

// Property has a name and a Value
type Property struct {
	Name  string `[ @(String) ":" ]`
	Value *Value `@@`
}

// Value represents any possible values type:
// - a String
// - a Null
// - an Array of value
// - a nested structure
//
// Different value types are a pointer to member of the type
type Value struct {
	Str     *string    `@(String)`
	Struct  *Structure `|  @@ `
	NullStr *string    `| "null" `
	Ar      []*Value   `| "[" { @@ {"," @@ } }"]"`
	Bool    *string    `|  @(Bool)|@(Property)`
	Number  *int       `| @(Number)`
}

// Define all kind of token to be recognized
// Structure characters {}[],:
// Property name, an UTF8 identifier followed by a column (:)
// "Quoted" and 'Apostrophe' strings with escaped quote
// null value
// and any space character \s
const jsLexer = `(\s+)|(?P<Bool>(true|false))|(?P<Property>[\p{L}_][\p{L}\d_]*)|(?P<String>'[^'\\]*(?:\\.[^'\\]*)*'|"[^"\\]*(?:\\.[^"\\]*)*")|(?P<Null>null)|(?P<Structure>[,|;{}\[\]:])|(?P<Number>(\d+(?:\.?\d+)?))`

//TODO: reuse jsLexer and jsParser

// ParseObject parse javascript object literal into a *Structure.
// Its returns a nil structure and an error when the object can't be parsed
func ParseObject(b []byte) (*Structure, error) {

	jsLexer := lexer.Must(lexer.Regexp(jsLexer))
	jsParser := participle.MustBuild(
		&Structure{},
		participle.Lexer(jsLexer),
		participle.Unquote(jsLexer, "String"),
		participle.UseLookahead(),
	)
	s := &Structure{}
	err := jsParser.ParseBytes(b, s)
	if err != nil {
		return nil, err
	}
	return s, nil
}
