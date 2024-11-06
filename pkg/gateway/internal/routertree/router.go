package routertree

import (
	"fmt"
	"strings"
	"sync/atomic"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/generic"

	"github.com/pubgo/lava/pkg/gateway/internal/routerparser"
)

var (
	ErrPathNodeNotFound = errors.New("path node not found")
	ErrNotFound         = errors.New("operation not found")
)

// RouteTree represents a prefix tree for routing
type RouteTree struct {
	root        *node
	cache       *lru.Cache[string, *MatchOperation]
	matchCount  int64
	cacheHits   int64
	cacheMisses int64
}

// node represents a node in the routing tree
type node struct {
	path     string
	children []*node
	isWild   bool // whether this is a wildcard node (* or **)
	target   *routeTarget
	indices  string // first letters of children paths for quick lookup
}

// routeTarget holds the endpoint information
type routeTarget struct {
	Method    string
	Path      string
	Operation string
	Verb      *string
	Vars      []*routerparser.PathVariable // reuse PathVariable from parser.go
	extras    map[string]any
}

// RouteOperation represents a route operation for external use
type RouteOperation struct {
	Method    string         `json:"method,omitempty"`
	Path      string         `json:"path,omitempty"`
	Operation string         `json:"operation,omitempty"`
	Verb      string         `json:"verb,omitempty"`
	Vars      []string       `json:"vars,omitempty"`
	Extras    map[string]any `json:"extras,omitempty"`
}

// MatchOperation represents a matched route operation
type MatchOperation struct {
	Method    string
	Path      string
	Operation string
	Verb      string
	Vars      []routerparser.PathFieldVar // reuse PathFieldVar from parser.go
	Extras    map[string]any
}

// NewRouteTree creates a new routing tree
func NewRouteTree() *RouteTree {
	cache := assert.Must1(lru.New[string, *MatchOperation](1000))
	return &RouteTree{
		root:  &node{},
		cache: cache,
	}
}

// Add adds a new route to the tree
func (r *RouteTree) Add(method string, path string, operation string, extras map[string]any) error {
	pattern, err := routerparser.ParseRoutePattern(path)
	if err != nil {
		return err
	}

	// 先添加 method 节点
	methodNode := r.root.findChild(handlerMethod(method))
	if methodNode == nil {
		methodNode = &node{path: handlerMethod(method)}
		r.root.children = append(r.root.children, methodNode)
		r.root.indices += string(methodNode.path[0])
	}

	n := methodNode
	// 移除空路径段
	for _, segment := range pattern.Segments {
		if segment == "" {
			continue
		}

		// 查找或创建子节点
		child := n.findChild(segment)
		if child == nil {
			child = &node{
				path:   segment,
				isWild: segment == routerparser.Star || segment == routerparser.DoubleStar,
			}
			n.children = append(n.children, child)
			n.indices += string(segment[0])
		}
		n = child
	}

	n.target = &routeTarget{
		Method:    method,
		Path:      path,
		Operation: operation,
		extras:    extras,
		Verb:      pattern.HttpVerb,
		Vars:      pattern.Variables,
	}

	return nil
}

// findChild finds a child node by path
func (n *node) findChild(path string) *node {
	for _, child := range n.children {
		if child.path == path {
			return child
		}
	}
	return nil
}

// Match finds a matching route for the given method and URL
func (r *RouteTree) Match(method, url string) (*MatchOperation, error) {
	atomic.AddInt64(&r.matchCount, 1)

	cacheKey := method + ":" + url
	if cached, ok := r.cache.Get(cacheKey); ok {
		atomic.AddInt64(&r.cacheHits, 1)
		return cached, nil
	}
	atomic.AddInt64(&r.cacheMisses, 1)

	result, err := r.match(method, url)
	if err == nil {
		r.cache.Add(cacheKey, result)
	}
	return result, err
}

// List returns all registered routes
func (r *RouteTree) List() []RouteOperation {
	ops := make([]RouteOperation, 0, 32)
	r.walkNode(r.root, "", &ops)
	return ops
}

// walkNode recursively walks the route tree and collects all operations
func (r *RouteTree) walkNode(n *node, prefix string, ops *[]RouteOperation) {
	if n == nil {
		return
	}

	// 构建当前路径
	path := prefix
	if len(n.path) > 0 {
		if path == "" {
			path = n.path
		} else {
			path += "/" + n.path
		}
	}

	// 如果当前节点有目标操作，添加到结果中
	if n.target != nil {
		vars := make([]string, 0, len(n.target.Vars))
		for _, v := range n.target.Vars {
			vars = append(vars, strings.Join(v.FieldPath, "."))
		}

		*ops = append(*ops, RouteOperation{
			Method:    n.target.Method,
			Path:      n.target.Path,
			Operation: n.target.Operation,
			Verb:      generic.FromPtr(n.target.Verb),
			Vars:      vars,
			Extras:    n.target.extras,
		})
	}

	// 递归遍历所有子节点
	for _, child := range n.children {
		r.walkNode(child, path, ops)
	}
}

// Stats returns the router's statistics
func (r *RouteTree) Stats() map[string]interface{} {
	return map[string]interface{}{
		"total_matches": atomic.LoadInt64(&r.matchCount),
		"cache_hits":    atomic.LoadInt64(&r.cacheHits),
		"cache_misses":  atomic.LoadInt64(&r.cacheMisses),
	}
}

// Helper functions
func handlerMethod(method string) string {
	return fmt.Sprintf("__%s__", strings.ToUpper(method))
}

// match finds a matching route
func (r *RouteTree) match(method, url string) (*MatchOperation, error) {
	// 先检查 method
	methodSegment := handlerMethod(method)

	// 解析URL路径
	paths := strings.Split(strings.Trim(url, "/"), "/")

	// 移除空路径段
	urlPaths := make([]string, 0, len(paths))
	for _, p := range paths {
		if p != "" {
			urlPaths = append(urlPaths, p)
		}
	}

	// 处理最后一个路径段中的verb
	verb := ""
	if len(urlPaths) > 0 {
		lastPath := urlPaths[len(urlPaths)-1]
		if idx := strings.LastIndex(lastPath, ":"); idx >= 0 {
			verb = lastPath[idx+1:]
			urlPaths[len(urlPaths)-1] = lastPath[:idx]
		}
	}

	// 执行匹配
	n := r.root
	var wildcardValues []string
	var doubleStarStart int = -1
	var doubleStarEnd int = -1

	// 检查是否存在对应的 method 节点
	var methodNode *node
	for _, child := range n.children {
		if child.path == methodSegment {
			methodNode = child
			break
		}
	}

	// 如果找不到对应的 method 节点，直接返回错误
	if methodNode == nil {
		return nil, ErrNotFound
	}
	n = methodNode

	// 匹配路径段
	for i := 0; i < len(urlPaths); i++ {
		path := urlPaths[i]
		found := false

		// 如果在双星号模式中
		if doubleStarStart >= 0 && doubleStarEnd == -1 {
			// 尝试查找下一个固定段
			for _, child := range n.children {
				if !child.isWild {
					// 找到固定段，记录双星号结束位置
					if child.path == path {
						doubleStarEnd = i
						n = child
						found = true
						break
					}
				}
			}
			if !found {
				// 继续收集双星号段
				continue
			}
		} else {
			// 常规匹配
			for _, child := range n.children {
				if !child.isWild && child.path == path {
					n = child
					found = true
					break
				} else if child.isWild {
					if child.path == routerparser.DoubleStar {
						doubleStarStart = i
						n = child
						found = true
						break
					} else if child.path == routerparser.Star {
						wildcardValues = append(wildcardValues, path)
						n = child
						found = true
						break
					}
				}
			}
		}

		if !found && doubleStarStart < 0 {
			return nil, ErrPathNodeNotFound
		}
	}

	// 确保找到了目标节点
	if n.target == nil {
		return nil, ErrNotFound
	}

	// 如果双星号匹配还未结束，设置结束位置为最后一个段
	if doubleStarStart >= 0 && doubleStarEnd == -1 {
		doubleStarEnd = len(urlPaths)
	}

	// 构建变量列表
	var vars []routerparser.PathFieldVar
	if len(n.target.Vars) > 0 {
		vars = make([]routerparser.PathFieldVar, len(n.target.Vars))
		wildcardIndex := 0

		for i, v := range n.target.Vars {
			if v.EndIdx == -1 && doubleStarStart >= 0 {
				// 处理双星号变量
				value := strings.Join(urlPaths[doubleStarStart:doubleStarEnd], "/")
				vars[i] = routerparser.PathFieldVar{
					Fields: v.FieldPath,
					Value:  value,
				}
			} else if v.EndIdx >= 0 && wildcardIndex < len(wildcardValues) {
				// 处理普通变量
				vars[i] = routerparser.PathFieldVar{
					Fields: v.FieldPath,
					Value:  wildcardValues[wildcardIndex],
				}
				wildcardIndex++
			}
		}
	}

	return &MatchOperation{
		Method:    method,
		Path:      n.target.Path,
		Operation: n.target.Operation,
		Verb:      verb,
		Vars:      vars,
		Extras:    n.target.extras,
	}, nil
}
