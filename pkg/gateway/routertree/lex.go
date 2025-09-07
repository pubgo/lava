package routertree

import (
	"github.com/alecthomas/participle/v2/lexer"
)

const (
	doubleStar = "**"
	star       = "*"
)

// httpRule
// Template = "/" Segments [ Verb ] ;
// Segments = Segment { "/" Segment } ;
// Segment  = "*" | "**" | LITERAL | Variable ;
// Variable = "{" FieldPath [ "=" Segments ] "}" ;
// FieldPath = IDENT { "." IDENT } ;
// Verb     = ":" LITERAL ;
type httpRule struct {
	Pos      lexer.Position
	Slash    string    `parser:"@\"/\""`
	Segments *segments `parser:"@@!"`
	Verb     *string   `parser:"(\":\" @Ident)?"`
}

// nolint
type segments struct {
	Pos      lexer.Position
	Segments []*segment `parser:"@@ (\"/\" @@)*"`
}

// nolint
type segment struct {
	Pos      lexer.Position
	Path     *string   `parser:"@(\"*\" \"*\" | \"*\" | Ident)"`
	Variable *variable `parser:"| @@*"`
}

// nolint
type variable struct {
	Pos      lexer.Position
	Fields   []string  `parser:"\"{\" @Ident (\".\" @Ident)*"`
	Segments *segments `parser:"(\"=\" @@)? \"}\""`
}
