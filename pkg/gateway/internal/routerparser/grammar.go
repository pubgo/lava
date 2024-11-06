package routerparser

import (
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/pkg/errors"
)

var (
	// 定义词法分析器
	routeLexer = lexer.MustSimple([]lexer.SimpleRule{
		{"whitespace", `\s+`},
		{"DoubleStar", `\*\*`}, // 必须在 Star 之前
		{"Star", `\*`},
		{"Slash", `/`},
		{"Colon", `:`},
		{"LBrace", `{`},
		{"RBrace", `}`},
		{"Equal", `=`},
		{"Dot", `\.`},
		{"Ident", `[a-zA-Z][a-zA-Z0-9_\-]*`}, // 标识符规则调整
	})

	// 创建解析器，使用正确的泛型语法
	routeParser = participle.MustBuild[RoutePattern](
		participle.Lexer(routeLexer),
		participle.Elide("whitespace"),
		participle.UseLookahead(2),
	)
)

// RoutePattern 表示完整的路由模式
type RoutePattern struct {
	Segments []*PathSegment `parser:"@@* ('/' @@)*"`
	Verb     *string        `parser:"(':' @Ident)?"`
}

// PathSegment 表示路径段
type PathSegment struct {
	Literal  *string   `parser:"  @Ident"`
	Variable *Variable `parser:"| '{' @@ '}'"`
}

// Variable 表示变量定义
type Variable struct {
	Name    string           `parser:"@(Ident ('.' Ident)*)"` // 支持带点的字段名
	Pattern *VariablePattern `parser:"('=' @@)?"`
}

// VariablePattern 表示变量模式
type VariablePattern struct {
	Parts []*PatternPart `parser:"@@ ('/' @@)*"`
}

// PatternPart 表示模式的组成部分
type PatternPart struct {
	Literal    *string `parser:"  @Ident"`
	Star       bool    `parser:"| @Star"`
	DoubleStar bool    `parser:"| @DoubleStar"`
}

// ParseRoute 解析路由模式
func ParseRoute(pattern string) (*RoutePattern, error) {
	if pattern == "" {
		return nil, errors.New("empty pattern")
	}

	// 移除开头的 /
	if pattern[0] == '/' {
		pattern = pattern[1:]
	}

	route, err := routeParser.ParseString("", pattern)
	if err != nil {
		return nil, errors.Wrap(err, "parse route failed")
	}

	// 验证 ** 的位置
	if err := validateDoubleStarPosition(route); err != nil {
		return nil, err
	}

	return route, nil
}

// validateDoubleStarPosition 验证 ** 的位置
func validateDoubleStarPosition(route *RoutePattern) error {
	hasDoubleStar := false
	for _, seg := range route.Segments {
		if seg.Variable != nil && seg.Variable.Pattern != nil {
			// 先检查是否有多个 **
			doubleStarCount := 0
			for _, part := range seg.Variable.Pattern.Parts {
				if part.DoubleStar {
					doubleStarCount++
					if doubleStarCount > 1 {
						return errors.New("multiple ** patterns are not allowed")
					}
				}
			}

			// 再检查 ** 的位置
			for i, part := range seg.Variable.Pattern.Parts {
				if part.DoubleStar {
					if hasDoubleStar {
						return errors.New("multiple ** patterns are not allowed")
					}
					if i != len(seg.Variable.Pattern.Parts)-1 {
						return errors.New("** must be the last part")
					}
					hasDoubleStar = true
				}
			}
		}
	}
	return nil
}

// String 返回路由模式的字符串表示
func (r *RoutePattern) String() string {
	var parts []string

	// 处理路径段
	for i, seg := range r.Segments {
		if i > 0 || len(r.Segments) == 0 {
			parts = append(parts, "/")
		}
		if seg.Literal != nil {
			parts = append(parts, *seg.Literal)
		} else if seg.Variable != nil {
			parts = append(parts, seg.Variable.String())
		}
	}

	// 处理动词
	if r.Verb != nil {
		parts = append(parts, ":", *r.Verb)
	}

	return strings.Join(parts, "")
}

// String 返回变量的字符串表示
func (v *Variable) String() string {
	if v.Pattern == nil {
		return "{" + v.Name + "}"
	}
	return "{" + v.Name + "=" + v.Pattern.String() + "}"
}

// String 返回变量模式的字符串表示
func (p *VariablePattern) String() string {
	var parts []string
	for i, part := range p.Parts {
		if i > 0 {
			parts = append(parts, "/")
		}
		parts = append(parts, part.String())
	}
	return strings.Join(parts, "")
}

// String 返回模式部分的字符串表示
func (p *PatternPart) String() string {
	if p.Literal != nil {
		return *p.Literal
	}
	if p.DoubleStar {
		return "**"
	}
	if p.Star {
		return "*"
	}
	return ""
}
