package routertree

import (
	"fmt"
	"strings"

	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/generic"
)

var (
	ErrPathNodeNotFound = errors.New("path node not found")
	ErrNotFound         = errors.New("operation not found")
)

func NewRouteTree() *RouteTree {
	return &RouteTree{nodes: make(map[string]*nodeTree)}
}

type RouteTree struct {
	nodes map[string]*nodeTree
}

func (r *RouteTree) List() []RouteOperation {
	return getOpt(r.nodes)
}

func (r *RouteTree) Add(method string, path string, operation string, extras map[string]any) error {
	var errMsg = func() string {
		return fmt.Sprintf("method: %s, path: %s, operation: %s", method, path, operation)
	}

	rule, err := parse(path)
	if err != nil {
		return errors.Wrap(err, errMsg())
	}

	var node = parseToRoute(rule)
	if len(node.Paths) == 0 {
		return errors.Wrap(fmt.Errorf("path is null"), errMsg())
	}

	var nodes = r.nodes
	paths := append(node.Paths, handlerMethod(method))
	for i, n := range paths {
		var lastNode = nodes[n]
		if lastNode == nil {
			lastNode = &nodeTree{nodes: make(map[string]*nodeTree), verbs: make(map[string]*routeTarget)}
			nodes[n] = lastNode
		}
		nodes = lastNode.nodes

		if i == len(paths)-1 {
			lastNode.verbs[generic.FromPtr(node.Verb)] = &routeTarget{
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
	var paths = strings.Split(strings.Trim(strings.TrimSpace(url), "/"), "/")
	var lastPath = strings.SplitN(paths[len(paths)-1], ":", 2)
	var errMsg = func(tags ...errors.Tag) errors.Tags {
		return append(tags, errors.T("method", method), errors.T("url", url))
	}
	var verb = ""

	paths[len(paths)-1] = lastPath[0]
	if len(lastPath) > 1 {
		verb = lastPath[1]
	}

	paths = append(paths, handlerMethod(method))
	var getVars = func(vars []*pathVariable, paths []string) []PathFieldVar {
		var vv = make([]PathFieldVar, 0, len(vars))
		for _, v := range vars {
			pathVar := PathFieldVar{Fields: v.Fields}
			if v.end > 0 {
				pathVar.Value = strings.Join(paths[v.start:v.end+1], "/")
			} else {
				pathVar.Value = strings.Join(paths[v.start:], "/")
			}

			vv = append(vv, pathVar)
		}
		return vv
	}
	var getPath = func(nodes map[string]*nodeTree, names ...string) *nodeTree {
		for _, n := range names {
			path := nodes[n]
			if path != nil {
				return path
			}
		}
		return nil
	}

	var nodes = r.nodes
	for _, n := range paths {
		path := getPath(nodes, n, star, doubleStar)
		if path == nil {
			return nil, errors.WrapFn(ErrPathNodeNotFound, func() errors.Tags {
				return errMsg(errors.T("node", n))
			})
		}

		if vv := path.verbs[verb]; vv != nil && vv.Operation != "" && vv.Method == method {
			return &MatchOperation{
				Extras:    vv.extras,
				Method:    vv.Method,
				Path:      vv.Path,
				Operation: vv.Operation,
				Verb:      verb,
				Vars:      getVars(vv.Vars, paths),
			}, nil
		}
		nodes = path.nodes
	}

	return nil, errors.WrapTag(ErrNotFound, errMsg()...)
}

type RouteOperation struct {
	Method    string         `json:"method,omitempty"`
	Path      string         `json:"path,omitempty"`
	Operation string         `json:"operation,omitempty"`
	Verb      string         `json:"verb,omitempty"`
	Vars      []string       `json:"vars,omitempty"`
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
	nodes map[string]*nodeTree
	verbs map[string]*routeTarget
}

type MatchOperation struct {
	Method    string
	Path      string
	Operation string
	Verb      string
	Vars      []PathFieldVar
	Extras    map[string]any
}

func getOpt(nodes map[string]*nodeTree) []RouteOperation {
	var sets []RouteOperation
	for _, n := range nodes {
		for _, v := range n.verbs {
			sets = append(sets, RouteOperation{
				Method:    v.Method,
				Path:      v.Path,
				Operation: v.Operation,
				Verb:      generic.FromPtr(v.Verb),
				Vars:      generic.Map(v.Vars, func(i int) string { return strings.Join(v.Vars[i].Fields, ".") }),
				Extras:    v.extras,
			})
		}
		sets = append(sets, getOpt(n.nodes)...)
	}
	return sets
}

func handlerMethod(method string) string {
	return fmt.Sprintf("__%s__", strings.ToUpper(method))
}
