package routerparser

import (
	"strings"

	"github.com/samber/lo"
)

// buildPattern 构建路由模式
func buildPattern(route *RoutePattern) (*Pattern, error) {
	p := &Pattern{
		raw:      "/" + route.String(),
		HttpVerb: route.Verb,
	}

	// 处理每个段
	startIdx := 0
	for _, seg := range route.Segments {
		if seg.Literal != nil {
			p.Segments = append(p.Segments, *seg.Literal)
			startIdx++
			continue
		}

		if seg.Variable != nil {
			if err := buildVariable(seg.Variable, startIdx, p); err != nil {
				return nil, err
			}
			startIdx += getVariableLength(seg.Variable)
		}
	}

	return p, nil
}

// getVariableLength 获取变量长度
func getVariableLength(v *Variable) int {
	if v.Pattern == nil {
		return 1
	}
	return len(v.Pattern.Parts)
}

// buildVariable 构建变量
func buildVariable(v *Variable, startIdx int, p *Pattern) error {
	fields := strings.Split(v.Name, ".")

	if v.Pattern == nil {
		// 简单变量
		p.Segments = append(p.Segments, Star)
		if p.Variables == nil {
			p.Variables = make([]*PathVariable, 0)
		}
		p.Variables = append(p.Variables, &PathVariable{
			FieldPath: fields,
			StartIdx:  startIdx,
			EndIdx:    startIdx,
		})
		return nil
	}

	// 复杂变量
	var patternParts []string
	hasDoubleStar := false

	for _, part := range v.Pattern.Parts {
		if part.DoubleStar {
			hasDoubleStar = true
			patternParts = append(patternParts, DoubleStar)
		} else if part.Star {
			patternParts = append(patternParts, Star)
		} else if part.Literal != nil {
			patternParts = append(patternParts, *part.Literal)
		}
	}

	if p.Variables == nil {
		p.Variables = make([]*PathVariable, 0)
	}
	p.Variables = append(p.Variables, &PathVariable{
		FieldPath: fields,
		StartIdx:  startIdx,
		EndIdx:    lo.Ternary(hasDoubleStar, -1, startIdx+len(patternParts)-1),
		Pattern:   strings.Join(patternParts, "/"),
	})

	// 添加所有段到 pattern
	p.Segments = append(p.Segments, patternParts...)
	return nil
}
