package routertree

import (
	"fmt"
	"strings"

	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/generic"
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

func (r *RouteTree) Add(method string, path string, operation string, extras map[string]any) error {
	errMsg := func() string {
		return fmt.Sprintf("method: %s, path: %s, operation: %s", method, path, operation)
	}

	rule, err := parse(path)
	if err != nil {
		return errors.Wrap(err, errMsg())
	}

	node := parseToRoute(rule)
	if len(node.Paths) == 0 {
		return errors.Format("node path is empty: %s", errMsg())
	}

	nodeMap := r.nodeMap
	method = handlerMethod(method)
	paths := node.Paths
	for i, n := range paths {
		var lastNode = nodeMap[n]
		if lastNode == nil {
			lastNode = &nodeTree{nodeMap: make(map[string]*nodeTree), verbMap: make(map[string]*routeTarget)}
			nodeMap[n] = lastNode
		}
		nodeMap = lastNode.nodeMap

		if i == len(paths)-1 {
			verbKey := fmt.Sprintf("%s:%s", method, generic.FromPtr(node.Verb))
			lastNode.verbMap[verbKey] = &routeTarget{
				Method:    method,
				Path:      path,
				Operation: operation,
				extras:    extras,
				Verb:      node.Verb,
				Vars:      node.Vars,
			}
		}
	}
	return nil
}

func (r *RouteTree) Match(method, url string) (*MatchOperation, error) {
	var pathNodes = strings.Split(strings.Trim(strings.TrimSpace(url), "/"), "/")
	var lastPath = strings.SplitN(pathNodes[len(pathNodes)-1], ":", 2)
	var errMsg = func(tags ...errors.Tag) errors.Tags {
		return append(tags, errors.T("method", method), errors.T("url", url))
	}
	var verb = ""

	pathNodes[len(pathNodes)-1] = lastPath[0]
	if len(lastPath) > 1 {
		verb = lastPath[1]
	}

	method = handlerMethod(method)
	verbKey := fmt.Sprintf("%s:%s", method, verb)

	var getVars = func(vars []*pathVariable, paths []string) []PathFieldVar {
		var vv = make([]PathFieldVar, 0, len(vars))
		for _, v := range vars {
			pathVar := PathFieldVar{Fields: v.fields}
			if v.end > 0 {
				pathVar.Value = strings.Join(paths[v.start:v.end+1], "/")
			} else {
				pathVar.Value = strings.Join(paths[v.start:], "/")
			}

			vv = append(vv, pathVar)
		}
		return vv
	}

	var getPath = func(nodeMap map[string]*nodeTree, names ...string) (string, *nodeTree) {
		for _, name := range names {
			path := nodeMap[name]
			if path != nil {
				return name, path
			}
		}
		return "", nil
	}

	nodeMap := r.nodeMap
	lastIndex := len(pathNodes) - 1
	for index, node := range pathNodes {
		nodeName, path := getPath(nodeMap, node, star, doubleStar)
		if path == nil {
			return nil, errors.WrapFn(ErrPathNodeNotFound, func() errors.Tags {
				return errMsg(errors.T("node", node))
			})
		}

		nodeMap = path.nodeMap
		switch nodeName {
		case node:
			if index != lastIndex {
				continue
			}
		case star:
			if index != lastIndex && len(path.nodeMap) != 0 {
				nextPath := path.nodeMap[pathNodes[index+1]]
				if nextPath != nil {
					continue
				}
			}
		case doubleStar:
		}

		vv := path.verbMap[verbKey]
		if vv == nil {
			return nil, errors.WrapTag(ErrOperationNotFound, errMsg(errors.T("node", node))...)
		}

		return &MatchOperation{
			Extras:    vv.extras,
			Method:    vv.Method,
			Path:      vv.Path,
			Operation: vv.Operation,
			Verb:      verb,
			Vars:      getVars(vv.Vars, pathNodes),
		}, nil
	}

	return nil, errors.WrapTag(ErrOperationNotFound, errMsg()...)
}

func getOpt(nodes map[string]*nodeTree) []RouteOperation {
	var sets []RouteOperation
	for _, n := range nodes {
		for _, v := range n.verbMap {
			sets = append(sets, RouteOperation{
				Method:    v.Method,
				Path:      v.Path,
				Operation: v.Operation,
				Verb:      generic.FromPtr(v.Verb),
				Vars:      generic.Map(v.Vars, func(i int) string { return strings.Join(v.Vars[i].fields, ".") }),
				Extras:    v.extras,
			})
		}
		sets = append(sets, getOpt(n.nodeMap)...)
	}
	return sets
}

func handlerMethod(method string) string {
	return fmt.Sprintf("__%s__", strings.ToUpper(method))
}
