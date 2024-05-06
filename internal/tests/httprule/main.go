package main

import (
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/pretty"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

var (
	lex = lexer.MustSimple([]lexer.SimpleRule{
		{"Ident", `[a-zA-Z]\w*`},
		{"Punct", `[-[!@#$%^&*()+_={}\|:;"'<,>.?/]|]`},
	})

	parser = participle.MustBuild[HttpRule](
		participle.Lexer(lex),
	)
)

//     Template = "/" Segments [ Verb ] ;
//     Segments = Segment { "/" Segment } ;
//     Segment  = "*" | "**" | LITERAL | Variable ;
//     Variable = "{" FieldPath [ "=" Segments ] "}" ;
//     FieldPath = IDENT { "." IDENT } ;
//     Verb     = ":" LITERAL ;
//

type HttpRule struct {
	Pos      lexer.Position
	Slash    string    `"/" `
	Segments *Segments `@@!`
	Verb     *string   `(":" @Ident)?`
}

type Segments struct {
	Pos      lexer.Position
	Segments []*Segment `@@ ("/" @@)*`
}

type Segment struct {
	Pos      lexer.Position
	Path     *string   `@("*" "*" | "*" | Ident)`
	Variable *Variable `| @@`
}

type Variable struct {
	Pos      lexer.Position
	Fields   []string  `"{" @Ident ("." @Ident)*`
	Segments *Segments `("=" @@)? "}"`
}

func main() {
	ini := assert.Must1(parser.Parse("",
		strings.NewReader("/v1/users/{aa.ba.ca=hh/*}/hello/{hello}/messages/{messageId=nn/ss/**}:change"),
		//participle.Trace(os.Stdout),
	))

	pretty.Println(ini)
}
