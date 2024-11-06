package routertree

import (
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/env"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/generic"
)

const (
	doubleStar = "**"
	star       = "*"
)

var (
	// Initialize LRU cache with a size of 1000
	// Adjust the size based on your needs
	ruleCache = assert.Must1(lru.New[string, *httpRule](1000))

	parser = assert.Exit1(participle.Build[httpRule](
		participle.Lexer(assert.Exit1(lexer.NewSimple([]lexer.SimpleRule{
			{Name: "Ident", Pattern: `[a-zA-Z][\w\_\-\.]*`},
			{Name: "Punct", Pattern: `[-[!@#$%^&*()+_={}\|:;"'<,>.?/]|]`},
		}))),
	))

	debug = env.Define("HTTP_RULE_PARSER_DEBUG", "Enable debug mode for HTTP rule parser").Bool()
)

// httpRule defines the syntax structure for HTTP routing rules
// Grammar:
// Template  = "/" Segments [ Verb ] ;
// Segments  = Segment { "/" Segment } ;
// Segment   = "*" | "**" | LITERAL | Variable ;
// Variable  = "{" FieldPath [ "=" Segments ] "}" ;
// FieldPath = IDENT { "." IDENT } ;
// Verb      = ":" LITERAL ;
type httpRule struct {
	Slash    string    `parser:"@'/'"`            // Must start with '/'
	Segments *segments `parser:"@@"`              // Parse path segments
	Verb     *string   `parser:"( ':' @Ident )?"` // Optional HTTP verb
}

type segments struct {
	Segments []*segment `parser:"@@ ( '/' @@ )*"` // One or more segments separated by '/'
}

type segment struct {
	Path     *string   `parser:"@( '*' '*' | '*' | Ident )"` // Path can be '**' or '*' or identifier
	Variable *variable `parser:"| @@"`                       // Or a variable
}

type variable struct {
	Field    string    `parser:"'{' @Ident ( '.' Ident )*"`
	Segments *segments `parser:"( '=' @@ )? '}'"`
}

type PathVariable struct {
	FieldPath []string
	StartIdx  int
	EndIdx    int
}

type RoutePattern struct {
	Segments  []string
	HttpVerb  *string
	Variables []*PathVariable
}

// PathFieldVar represents a parsed path variable with its field path and value
type PathFieldVar struct {
	Fields []string
	Value  string
}

func (r *RoutePattern) Match(urls []string, verb string) ([]PathFieldVar, error) {
	if len(urls) < len(r.Segments) {
		return nil, errors.Format("url segments length %d is less than path length %d", len(urls), len(r.Segments))
	}

	if r.HttpVerb != nil && generic.FromPtr(r.HttpVerb) != verb {
		return nil, errors.New("verb not match")
	}

	// 提取变量值
	var vars []PathFieldVar
	lastSegmentIndex := len(urls) - 1

	for _, v := range r.Variables {
		var value string

		if v.EndIdx == -1 {
			// 双星号变量：收集从 StartIdx 到最后一个固定段之前的所有内容
			endIndex := lastSegmentIndex
			// 查找下一个固定段的位置
			for i := v.StartIdx + 1; i < len(r.Segments); i++ {
				if r.Segments[i] != star && r.Segments[i] != doubleStar {
					// 从后向前查找匹配的固定段
					for j := lastSegmentIndex; j > v.StartIdx; j-- {
						if urls[j] == r.Segments[i] {
							endIndex = j - 1
							break
						}
					}
					break
				}
			}

			if endIndex >= v.StartIdx {
				segments := urls[v.StartIdx : endIndex+1]
				value = strings.Join(segments, "/")
			}
		} else {
			// 普通变量
			if v.EndIdx >= v.StartIdx {
				value = strings.Join(urls[v.StartIdx:v.EndIdx+1], "/")
			} else {
				value = urls[v.StartIdx]
			}
		}

		vars = append(vars, PathFieldVar{
			Fields: v.FieldPath,
			Value:  value,
		})
	}

	// 验证固定路径段
	for i, segment := range r.Segments {
		if segment == star || segment == doubleStar {
			continue
		}

		// 找到对应的 URL 段进行比较
		matched := false
		for j := i; j < len(urls); j++ {
			if urls[j] == segment {
				matched = true
				break
			}
		}

		if !matched {
			return nil, errors.New("path not match")
		}
	}

	return vars, nil
}

// String generates a string representation of the RoutePattern
func (r *RoutePattern) String() string {
	url := "/"

	paths := make([]string, len(r.Segments))
	copy(paths, r.Segments)

	// Sort variables by StartIdx to ensure consistent output
	sortedVars := make([]*PathVariable, len(r.Variables))
	copy(sortedVars, r.Variables)
	sort.Slice(sortedVars, func(i, j int) bool {
		return sortedVars[i].StartIdx < sortedVars[j].StartIdx
	})

	for _, v := range sortedVars {
		varS := "{" + strings.Join(v.FieldPath, ".")

		// 处理双星号的情况
		if v.EndIdx == -1 {
			varS += "=**"
		} else {
			varS += "=" + paths[v.StartIdx]
			for i := v.StartIdx + 1; i <= v.EndIdx; i++ {
				varS += "/" + paths[i]
			}
		}

		varS += "}"
		paths[v.StartIdx] = varS

		// 清除已处理的路径段
		if v.EndIdx == -1 {
			// 对于双星号，清除到最后一个段之前
			for i := v.StartIdx + 1; i < len(paths)-1; i++ {
				paths[i] = ""
			}
		} else {
			// 对于普通变量，只清除变量范围内的段
			for i := v.StartIdx + 1; i <= v.EndIdx; i++ {
				paths[i] = ""
			}
		}
	}

	// 过滤掉空的路径段
	var segments []string
	for _, p := range paths {
		if p != "" {
			segments = append(segments, p)
		}
	}
	url += strings.Join(segments, "/")

	if r.HttpVerb != nil {
		url += ":" + generic.FromPtr(r.HttpVerb)
	}

	return url
}

func handleSegments(s *segment, rr *RoutePattern) {
	if s.Path != nil {
		rr.Segments = append(rr.Segments, *s.Path)
		return
	}

	fields := strings.Split(s.Variable.Field, ".")

	vv := &PathVariable{
		FieldPath: fields,
		StartIdx:  len(rr.Segments),
	}

	if s.Variable.Segments == nil {
		rr.Segments = append(rr.Segments, star)
	} else {
		for _, v := range s.Variable.Segments.Segments {
			handleSegments(v, rr)
		}
	}

	if len(rr.Segments) > 0 && rr.Segments[len(rr.Segments)-1] == doubleStar {
		vv.EndIdx = -1
	} else {
		vv.EndIdx = len(rr.Segments) - 1
	}

	rr.Variables = append(rr.Variables, vv)
}

func parseToRoute(rule *httpRule) *RoutePattern {
	r := new(RoutePattern)
	r.HttpVerb = rule.Verb

	if rule.Segments != nil {
		for _, v := range rule.Segments.Segments {
			handleSegments(v, r)
		}
	}

	return r
}

func parse(url string) (*httpRule, error) {
	// Try to get from cache
	if cached, ok := ruleCache.Get(url); ok {
		return cached, nil
	}

	var options = make([]participle.ParseOption, 0)
	if debug {
		options = append(options, participle.Trace(os.Stdout))
		options = append(options, participle.AllowTrailing(true))
	}

	// Parse if not in cache
	rule, err := parser.ParseString("", url, options...)
	if err != nil {
		return nil, err
	}

	// Add to cache
	ruleCache.Add(url, rule)
	return rule, nil
}

// Equal compares two RoutePatterns for equality
func (r *RoutePattern) Equal(other *RoutePattern) bool {
	if r == nil || other == nil {
		return r == other
	}

	// Compare segments
	if !reflect.DeepEqual(r.Segments, other.Segments) {
		return false
	}

	// Compare HttpVerb
	if !reflect.DeepEqual(r.HttpVerb, other.HttpVerb) {
		return false
	}

	// Compare Variables
	if len(r.Variables) != len(other.Variables) {
		return false
	}

	// Sort variables by StartIdx for consistent comparison
	rVars := make([]*PathVariable, len(r.Variables))
	otherVars := make([]*PathVariable, len(other.Variables))
	copy(rVars, r.Variables)
	copy(otherVars, other.Variables)

	sort.Slice(rVars, func(i, j int) bool {
		return rVars[i].StartIdx < rVars[j].StartIdx
	})
	sort.Slice(otherVars, func(i, j int) bool {
		return otherVars[i].StartIdx < otherVars[j].StartIdx
	})

	for i := range rVars {
		if !reflect.DeepEqual(rVars[i].FieldPath, otherVars[i].FieldPath) {
			return false
		}
		if rVars[i].StartIdx != otherVars[i].StartIdx {
			return false
		}
		if rVars[i].EndIdx != otherVars[i].EndIdx {
			return false
		}
	}

	return true
}
