//nolint:all
package routertree

import (
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/generic"
)

const (
	doubleStar = "**"
	star       = "*"
)

var (
	parser = assert.Exit1(participle.Build[httpRule](
		participle.Lexer(assert.Exit1(lexer.NewSimple([]lexer.SimpleRule{
			{Name: "Ident", Pattern: `[a-zA-Z][\w\_\-\.]*`},
			{Name: "Punct", Pattern: `[-[!@#$%^&*()+_={}\|:;"'<,>.?/]|]`},
		}))),
	))
)

// httpRule
// Template = "/" Segments [ Verb ] ;
// Segments = Segment { "/" Segment } ;
// Segment  = "*" | "**" | LITERAL | Variable ;
// Variable = "{" FieldPath [ "=" Segments ] "}" ;
// FieldPath = IDENT { "." IDENT } ;
// Verb     = ":" LITERAL ;
type httpRule struct {
	Slash    string    `@"/"`
	Segments *segments `@@!`
	Verb     *string   `(":" @Ident)?`
}

type segments struct {
	Segments []*segment `@@ ("/" @@)*`
}

type segment struct {
	Path     *string   `@("*" "*" | "*" | Ident)`
	Variable *variable `| @@*`
}

type variable struct {
	Fields   []string  `"{" @Ident ("." @Ident)*`
	Segments *segments `("=" @@)? "}"`
}

type pathVariable struct {
	Fields     []string
	start, end int
}

type routePath struct {
	Paths []string
	Verb  *string
	Vars  []*pathVariable
}

type PathFieldVar struct {
	Fields []string
	Value  string
}

func (r routePath) Match(urls []string, verb string) ([]PathFieldVar, error) {
	if len(urls) < len(r.Paths) {
		return nil, errors.New("urls length not match")
	}

	if r.Verb != nil {
		if generic.FromPtr(r.Verb) != verb {
			return nil, errors.New("verb not match")
		}
	}

	for i := range r.Paths {
		path := r.Paths[i]
		if path == star {
			continue
		}

		if path == urls[i] {
			continue
		}

		if path == doubleStar {
			continue
		}

		return nil, errors.New("path is not match")
	}

	var vv []PathFieldVar
	for _, v := range r.Vars {
		pathVar := PathFieldVar{Fields: v.Fields}
		if v.end > 0 {
			pathVar.Value = strings.Join(urls[v.start:v.end+1], "/")
		} else {
			pathVar.Value = strings.Join(urls[v.start:], "/")
		}

		vv = append(vv, pathVar)
	}

	return vv, nil
}

func (r routePath) String() string {
	url := "/"

	paths := make([]string, len(r.Paths))
	copy(paths, r.Paths)

	for _, v := range r.Vars {
		varS := "{" + strings.Join(v.Fields, ".") + "="
		end := generic.Ternary(v.end == -1, len(paths)-1, v.end)

		for i := v.start; i <= end; i++ {
			varS += generic.Ternary(i == v.start, paths[i], "/"+paths[i])
			if i > v.start {
				paths[i] = ""
			}
		}

		varS += "}"
		paths[v.start] = varS
	}

	url += strings.Join(generic.Filter(paths, func(s string) bool { return s != "" }), "/")

	if r.Verb != nil {
		url += ":" + generic.FromPtr(r.Verb)
	}

	return url
}

func handleSegments(s *segment, rr *routePath) {
	if s.Path != nil {
		rr.Paths = append(rr.Paths, *s.Path)
		return
	}

	vv := &pathVariable{Fields: s.Variable.Fields, start: len(rr.Paths)}
	if s.Variable.Segments == nil {
		rr.Paths = append(rr.Paths, star)
	} else {
		for _, v := range s.Variable.Segments.Segments {
			handleSegments(v, rr)
		}
	}

	if len(rr.Paths) > 0 && rr.Paths[len(rr.Paths)-1] == doubleStar {
		vv.end = -1
	} else {
		vv.end = len(rr.Paths) - 1
	}

	rr.Vars = append(rr.Vars, vv)
}

func parseToRoute(rule *httpRule) *routePath {
	r := new(routePath)
	r.Verb = rule.Verb

	if rule.Segments != nil {
		for _, v := range rule.Segments.Segments {
			handleSegments(v, r)
		}
	}

	return r
}

func parse(url string) (*httpRule, error) {
	return parser.ParseString("", url)
	//participle.AllowTrailing(true),
	//participle.Trace(os.Stdout),
}
