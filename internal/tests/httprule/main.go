package main

import (
	"strings"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/pretty"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

var (
	lex = lexer.MustSimple([]lexer.SimpleRule{
		{Name: "Ident", Pattern: `[a-zA-Z]\w*`},
		{Name: "Punct", Pattern: `[-[!@#$%^&*()+_={}\|:;"'<,>.?/]|]`},
	})

	parser = participle.MustBuild[HttpRule](
		participle.Lexer(lex),
	)
)

//     http rule
//     Template = "/" Segments [ Verb ] ;
//     Segments = Segment { "/" Segment } ;
//     Segment  = "*" | "**" | LITERAL | Variable ;
//     Variable = "{" FieldPath [ "=" Segments ] "}" ;
//     FieldPath = IDENT { "." IDENT } ;
//     Verb     = ":" LITERAL ;

type HttpRule struct {
	Pos      lexer.Position
	Slash    string    `"/"`
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

type pathVariable struct {
	Fields     []string
	start, end int
}

type RouteTarget struct {
	Paths []string
	Verb  *string
	Vars  []*pathVariable
}

func handleSegments(s *Segment, rr *RouteTarget) {
	if s.Path != nil {
		rr.Paths = append(rr.Paths, *s.Path)
		return
	}

	vv := &pathVariable{Fields: s.Variable.Fields, start: len(rr.Paths)}
	if s.Variable.Segments == nil {
		rr.Paths = append(rr.Paths, "*")
	} else {
		for _, v := range s.Variable.Segments.Segments {
			handleSegments(v, rr)
		}
	}

	vv.end = len(rr.Paths) - 1
	if len(rr.Paths) > 0 && rr.Paths[len(rr.Paths)-1] == "**" {
		vv.end = -1
	}

	rr.Vars = append(rr.Vars, vv)
}

func Eval(rule *HttpRule) *RouteTarget {
	var r = new(RouteTarget)
	r.Verb = rule.Verb

	if rule.Segments != nil {
		for _, v := range rule.Segments.Segments {
			handleSegments(v, r)
		}
	}

	return r
}

func main() {
	ini := assert.Must1(parser.Parse("",
		strings.NewReader("/v1/users/{aa.ba.ca=hh/*}/hello/{hello.abc}/messages/{messageId=nn/ss/**}:change"),
		// participle.Trace(os.Stdout),
	))

	pretty.Println(ini)

	pretty.Println(Eval(ini))
}
