package routertree

import (
	"fmt"
	"strings"

	"github.com/pubgo/funk/v2"
	"github.com/pubgo/funk/v2/errors"
	"github.com/samber/lo"
)

var (
	ErrPathNodeNotFound  = errors.New("path node not found")
	ErrOperationNotFound = errors.New("operation not found")
)

func New() *RouteTree {
	return &RouteTree{nodeMap: make(map[string]*nodeTree)}
}

type RouteOperation struct {
	Method    string         `json:"method,omitempty"`
	Path      string         `json:"path,omitempty"`
	Operation string         `json:"operation,omitempty"`
	Verb      string         `json:"verb,omitempty"`
	Vars      []string       `json:"vars,omitempty"`
	Extras    map[string]any `json:"extras"`
}

type MatchOperation struct {
	Method    string         `json:"method"`
	Path      string         `json:"path"`
	Operation string         `json:"operation"`
	Verb      string         `json:"verb"`
	Vars      []PathFieldVar `json:"vars"`
	Extras    map[string]any `json:"extras"`
}

type routeTarget struct {
	Method    string
	Path      string
	Operation string
	Verb      *string
	Vars      []*pathVariable
	extras    map[string]any
}

type nodeTree struct {
	nodeMap map[string]*nodeTree
	verbMap map[string]*routeTarget
}

type RouteTree struct {
	nodeMap map[string]*nodeTree
}

func (r *RouteTree) List() []RouteOperation {
	return getOpt(r.nodeMap)
}

func (r *RouteTree) Add(method, path, operation string, extras map[string]any) error {
	// 验证输入
	if path == "" {
		return errors.New("path cannot be empty")
	}
	if operation == "" {
		return errors.New("operation cannot be empty")
	}

	errMsg := func() string {
		return fmt.Sprintf("method: %s, path: %s, operation: %s", method, path, operation)
	}

	rule, err := parse(path)
	if err != nil {
		return errors.Wrap(err, errMsg())
	}

	node := parseToRoute(rule)
	method = handlerMethod(method)

	// 特殊处理根路径 "/"（空路径）
	if len(node.Paths) == 0 {
		rootNode := r.nodeMap[rootPathKey]
		if rootNode == nil {
			rootNode = &nodeTree{
				nodeMap: make(map[string]*nodeTree),
				verbMap: make(map[string]*routeTarget),
			}
			r.nodeMap[rootPathKey] = rootNode
		}
		return r.registerRoute(rootNode, method, path, operation, node.Verb, node.Vars, extras)
	}

	// 普通路径：遍历路径段，构建路由树
	nodeMap := r.nodeMap
	for i, pathSegment := range node.Paths {
		lastNode := nodeMap[pathSegment]
		if lastNode == nil {
			lastNode = &nodeTree{
				nodeMap: make(map[string]*nodeTree),
				verbMap: make(map[string]*routeTarget),
			}
			nodeMap[pathSegment] = lastNode
		}

		// 如果是最后一个路径段，注册路由
		if i == len(node.Paths)-1 {
			if err := r.registerRoute(lastNode, method, path, operation, node.Verb, node.Vars, extras); err != nil {
				return err
			}
		}

		nodeMap = lastNode.nodeMap
	}
	return nil
}

// matchNode 在指定节点上查找匹配的路由目标
// 返回匹配到的路由目标，如果匹配失败返回 nil
func (r *RouteTree) matchNode(node *nodeTree, verbKey string) *routeTarget {
	if node == nil {
		return nil
	}
	return node.verbMap[verbKey]
}

func (r *RouteTree) Match(method, url string) (*MatchOperation, error) {
	// 解析 URL
	pathNodes, verb, err := parseURL(url)
	if err != nil {
		return nil, errors.WrapTags(err, errors.Tags{"method": method, "url": url})
	}

	method = handlerMethod(method)
	verbKey := fmt.Sprintf("%s:%s", method, verb)

	errMsg := func(key string, value any) errors.Tags {
		tt := errors.Tags{"method": method, "url": url}
		if key != "" {
			tt[key] = value
		}
		return tt
	}

	// 特殊处理根路径 "/"
	if len(pathNodes) == 0 {
		if rootNode := r.nodeMap[rootPathKey]; rootNode != nil {
			if target := r.matchNode(rootNode, verbKey); target != nil {
				return buildMatchOperation(target, verb, pathNodes), nil
			}
		}
		return nil, errors.WrapTags(ErrOperationNotFound, errMsg("", nil))
	}

	// 递归匹配函数：匹配策略优先级为 精确匹配 > * 通配符 > ** 通配符
	var matchPath func(nodeMap map[string]*nodeTree, pathIndex int) (*MatchOperation, error)
	matchPath = func(nodeMap map[string]*nodeTree, pathIndex int) (*MatchOperation, error) {
		if pathIndex >= len(pathNodes) {
			return nil, errors.WrapTags(ErrOperationNotFound, errMsg("", nil))
		}

		pathSegment := pathNodes[pathIndex]
		isLast := pathIndex == len(pathNodes)-1

		// 1. 尝试精确匹配
		if exactNode := nodeMap[pathSegment]; exactNode != nil {
			if isLast {
				if target := r.matchNode(exactNode, verbKey); target != nil {
					return buildMatchOperation(target, verb, pathNodes), nil
				}
			} else {
				// 递归继续匹配下一个路径段
				if result, err := matchPath(exactNode.nodeMap, pathIndex+1); err == nil {
					return result, nil
				}
				// 精确匹配失败，继续尝试通配符
			}
		}

		// 2. 尝试 * 通配符（匹配单个路径段）
		if wildcardNode := nodeMap[star]; wildcardNode != nil {
			if isLast {
				if target := r.matchNode(wildcardNode, verbKey); target != nil {
					return buildMatchOperation(target, verb, pathNodes), nil
				}
			} else {
				// 递归继续匹配下一个路径段
				if result, err := matchPath(wildcardNode.nodeMap, pathIndex+1); err == nil {
					return result, nil
				}
			}
		}

		// 3. 尝试 ** 通配符（贪婪匹配所有剩余路径段）
		if doubleWildcardNode := nodeMap[doubleStar]; doubleWildcardNode != nil {
			if target := r.matchNode(doubleWildcardNode, verbKey); target != nil {
				return buildMatchOperation(target, verb, pathNodes), nil
			}
		}

		return nil, errors.WrapTags(ErrPathNodeNotFound, errMsg("node", pathSegment))
	}

	return matchPath(r.nodeMap, 0)
}

func getOpt(nodes map[string]*nodeTree) []RouteOperation {
	var sets []RouteOperation
	for _, n := range nodes {
		for _, v := range n.verbMap {
			sets = append(sets, RouteOperation{
				Method:    v.Method,
				Path:      v.Path,
				Operation: v.Operation,
				Verb:      lo.FromPtr(v.Verb),
				Vars:      funk.Map(v.Vars, func(v *pathVariable) string { return strings.Join(v.fields, ".") }),
				Extras:    v.extras,
			})
		}
		sets = append(sets, getOpt(n.nodeMap)...)
	}
	return sets
}

const (
	methodPrefix = "__"
	methodSuffix = "__"
	rootPathKey  = "" // 根路径 "/" 在 nodeMap 中使用空字符串作为键
)

func handlerMethod(method string) string {
	return methodPrefix + strings.ToUpper(method) + methodSuffix
}

// parseURL 解析 URL，返回路径节点列表和动词
// 区分空路径 "" 和根路径 "/"，空路径返回错误
func parseURL(url string) (pathNodes []string, verb string, err error) {
	originalURL := url
	url = strings.TrimSpace(url)
	trimmedURL := strings.Trim(url, "/")

	if trimmedURL == "" {
		// 区分空路径 "" 和根路径 "/"
		if originalURL == "" || strings.TrimSpace(originalURL) == "" {
			return nil, "", errors.WrapTags(ErrPathNodeNotFound, errors.Tags{
				"url": originalURL,
			})
		}
		// 根路径 "/"
		return []string{}, "", nil
	}

	pathNodes = strings.Split(trimmedURL, "/")

	// 分离动词和路径（格式：path:verb）
	if len(pathNodes) > 0 {
		lastPath := strings.SplitN(pathNodes[len(pathNodes)-1], ":", 2)
		if len(lastPath) > 1 {
			verb = lastPath[1]
		}
		pathNodes[len(pathNodes)-1] = lastPath[0]
	}

	return pathNodes, verb, nil
}

// extractPathVars 从路径变量中提取值
func extractPathVars(vars []*pathVariable, paths []string) []PathFieldVar {
	vv := make([]PathFieldVar, 0, len(vars))
	for _, v := range vars {
		pathVar := PathFieldVar{Fields: v.fields}
		// 边界检查
		if v.end > 0 && v.end < len(paths) {
			pathVar.Value = strings.Join(paths[v.start:v.end+1], "/")
		} else if v.start < len(paths) {
			pathVar.Value = strings.Join(paths[v.start:], "/")
		} else {
			pathVar.Value = ""
		}
		vv = append(vv, pathVar)
	}
	return vv
}

// buildMatchOperation 构建匹配结果
func buildMatchOperation(target *routeTarget, verb string, pathNodes []string) *MatchOperation {
	return &MatchOperation{
		Extras:    target.extras,
		Method:    target.Method,
		Path:      target.Path,
		Operation: target.Operation,
		Verb:      verb,
		Vars:      extractPathVars(target.Vars, pathNodes),
	}
}

// registerRoute 注册路由到指定节点的 verbMap
func (r *RouteTree) registerRoute(node *nodeTree, method, path, operation string, verb *string, vars []*pathVariable, extras map[string]any) error {
	verbKey := fmt.Sprintf("%s:%s", method, lo.FromPtr(verb))

	// 检查路由是否已存在
	if existing, exists := node.verbMap[verbKey]; exists {
		return errors.Errorf("route already exists: method=%s path=%s operation=%s (existing operation: %s)",
			method, path, operation, existing.Operation)
	}

	node.verbMap[verbKey] = &routeTarget{
		Method:    method,
		Path:      path,
		Operation: operation,
		extras:    extras,
		Verb:      verb,
		Vars:      vars,
	}
	return nil
}
