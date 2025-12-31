package routertree

import (
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/samber/lo"

	"github.com/pubgo/funk/v2"
	"github.com/pubgo/funk/v2/errors"
)

var parser = participle.MustBuild[httpRule](
	participle.Lexer(lexer.MustSimple([]lexer.SimpleRule{
		{Name: "Ident", Pattern: `[a-zA-Z][\w\_\-\.]*`},
		{Name: "Punct", Pattern: `[-[!@#$%^&*()+_={}\|:;"'<,>.?/]|]`},
	})),
)

type pathVariable struct {
	fields     []string
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
		if lo.FromPtr(r.Verb) != verb {
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

		return nil, errors.Errorf("path(%s) not match", path)
	}

	vv := make([]PathFieldVar, 0, len(r.Vars))
	for _, v := range r.Vars {
		pathVar := PathFieldVar{Fields: v.fields}
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
		varS := "{" + strings.Join(v.fields, ".") + "="
		end := funk.Ternary(v.end == -1, len(paths)-1, v.end)

		for i := v.start; i <= end; i++ {
			varS += funk.Ternary(i == v.start, paths[i], "/"+paths[i])
			if i > v.start {
				paths[i] = ""
			}
		}

		varS += "}"
		paths[v.start] = varS
	}

	url += strings.Join(funk.Filter(paths, func(s string) bool { return s != "" }), "/")

	if r.Verb != nil {
		url += ":" + lo.FromPtr(r.Verb)
	}

	return url
}

func handleSegments(s *segment, rr *routePath) {
	if s.Path != nil {
		rr.Paths = append(rr.Paths, *s.Path)
		return
	}

	vv := &pathVariable{fields: s.Variable.Fields, start: len(rr.Paths)}
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
	return parser.ParseString(
		"",
		url,
		// participle.AllowTrailing(true),
		// participle.Trace(os.Stdout),
	)
}
