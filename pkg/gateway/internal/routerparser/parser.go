package routerparser

import (
	"strings"

	"github.com/pkg/errors"
)

const (
	Star       = "*"
	DoubleStar = "**"
)

// Pattern 表示解析后的路由模式
type Pattern struct {
	Raw       string
	HttpVerb  *string
	Segments  []string
	Variables []*PathVariable
}

// PathVariable 表示路径变量
type PathVariable struct {
	FieldPath []string
	StartIdx  int
	EndIdx    int
	Pattern   string
}

// PathFieldVar 表示匹配到的变量值
type PathFieldVar struct {
	Fields []string
	Value  string
}

// ParsePattern 解析路由模式
func ParsePattern(pattern string) (*Pattern, error) {
	if pattern == "" {
		return nil, errors.New("empty pattern")
	}

	// 使用 participle 解析
	route, err := ParseRoute(pattern)
	if err != nil {
		return nil, errors.Wrap(err, "parse route failed")
	}

	return buildPattern(route)
}

// Match 匹配路由
func (p *Pattern) Match(urls []string, verb string) ([]PathFieldVar, error) {
	// 检查动词
	if p.HttpVerb != nil && verb != *p.HttpVerb {
		return nil, errors.New("verb not match")
	}

	// 检查固定段
	for i, segment := range p.Segments {
		if i >= len(urls) {
			return nil, errors.New("url segments too short")
		}

		// 如果不是通配符，需要严格匹配
		if segment != Star && segment != DoubleStar {
			if urls[i] != segment {
				return nil, errors.Errorf("segment not match, expect %s got %s", segment, urls[i])
			}
		}
	}

	// 提取变量值
	var result []PathFieldVar
	for _, v := range p.Variables {
		value, err := p.extractValue(v, urls)
		if err != nil {
			return nil, err
		}
		result = append(result, PathFieldVar{
			Fields: v.FieldPath,
			Value:  value,
		})
	}

	return result, nil
}

// extractValue 提取变量值
func (p *Pattern) extractValue(v *PathVariable, urls []string) (string, error) {
	if v.StartIdx >= len(urls) {
		return "", errors.New("url segments too short")
	}

	// 如果变量没有模式，直接返回单个值
	if v.Pattern == "" {
		return urls[v.StartIdx], nil
	}

	// 找到当前变量在 pattern 中的位置
	var currentIndex int
	for i, variable := range p.Variables {
		if variable == v {
			currentIndex = i
			break
		}
	}

	// 解析变量模式
	patternParts := strings.Split(v.Pattern, "/")
	urlIndex := v.StartIdx
	var segments []string

	// 找到变量的边界
	endIndex := len(urls)
	if currentIndex < len(p.Variables)-1 {
		endIndex = p.Variables[currentIndex+1].StartIdx
	}

	// 检查是否是 ** 模式
	if len(patternParts) > 0 && patternParts[len(patternParts)-1] == DoubleStar {
		// 验证 ** 是最后一个部分
		if len(patternParts) > 1 {
			// 先匹配前面的固定部分
			for _, part := range patternParts[:len(patternParts)-1] {
				if urlIndex >= len(urls) {
					return "", errors.New("url segments too short")
				}
				if part == Star {
					segments = append(segments, urls[urlIndex])
				} else if urls[urlIndex] != part {
					return "", errors.Errorf("segment not match, expect %s got %s", part, urls[urlIndex])
				}
				urlIndex++
			}
		}

		// 处理 ** 部分
		if currentIndex == len(p.Variables)-1 {
			// 如果是最后一个变量，使用所有剩余段
			remaining := urls[urlIndex:]
			if len(remaining) > 0 {
				segments = append(segments, strings.Join(remaining, "/"))
			}
		} else {
			// 找到下一个变量的起始位置
			nextVarStart := p.Variables[currentIndex+1].StartIdx
			if urlIndex < nextVarStart {
				segments = append(segments, strings.Join(urls[urlIndex:nextVarStart], "/"))
			}
		}
	} else {
		// 处理普通模式（固定段和 * 通配符）
		for _, part := range patternParts {
			// 先检查 URL 长度
			if urlIndex >= len(urls) {
				return "", errors.New("url segments too short")
			}

			// 检查固定段匹配
			if part != Star {
				if urls[urlIndex] != part {
					return "", errors.Errorf("segment not match, expect %s got %s", part, urls[urlIndex])
				}
			} else {
				segments = append(segments, urls[urlIndex])
			}
			urlIndex++

			// 检查边界
			if urlIndex > endIndex {
				return "", errors.New("url segments exceed variable boundary")
			}
		}
	}

	// 过滤掉空段
	var result []string
	for _, s := range segments {
		if s != "" {
			result = append(result, s)
		}
	}

	return strings.Join(result, "/"), nil
}
