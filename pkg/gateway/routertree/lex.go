package routertree

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
	Slash    string    `parser:"@\"/\""`
	Segments *segments `parser:"@@!"`
	Verb     *string   `parser:"(\":\" @Ident)?"`
}

// nolint
type segments struct {
	Segments []*segment `parser:"@@ (\"/\" @@)*"`
}

// nolint
type segment struct {
	Path     *string   `parser:"@(\"*\" \"*\" | \"*\" | Ident)"`
	Variable *variable `parser:"| @@*"`
}

// nolint
type variable struct {
	Fields   []string  `parser:"\"{\" @Ident (\".\" @Ident)*"`
	Segments *segments `parser:"(\"=\" @@)? \"}\""`
}
