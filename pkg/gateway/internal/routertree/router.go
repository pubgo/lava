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
	return &RouteTree{nodes: make(map[string]*pathNode)}
}

type RouteTree struct {
	nodes map[string]*pathNode
}

func (r *RouteTree) Add(method string, url string, operation string) error {
	rule, err := parse(url)
	if err != nil {
		return err
	}

	var node = parseToRoute(rule)
	if len(node.Paths) == 0 {
		return fmt.Errorf("path is null")
	}

	var nodes = r.nodes
	for i, n := range node.Paths {
		var lastNode = nodes[n]
		if lastNode == nil {
			lastNode = &pathNode{
				nodes: make(map[string]*pathNode),
				verbs: make(map[string]*routeTarget),
			}
			nodes[n] = lastNode
		}
		nodes = lastNode.nodes

		if i == len(node.Paths)-1 {
			lastNode.verbs[generic.FromPtr(node.Verb)] = &routeTarget{
				Method:    method,
				Operation: &operation,
				Verb:      &method,
				Vars:      node.Vars,
			}
		}
	}
	return nil
}

func (r *RouteTree) Match(method, url string) (*MatchOperation, error) {
	var paths = strings.Split(strings.Trim(strings.TrimSpace(url), "/"), "/")
	var lastPath = strings.SplitN(paths[len(paths)-1], ":", 2)
	var verb = ""

	paths[len(paths)-1] = lastPath[0]
	if len(lastPath) > 1 {
		verb = lastPath[1]
	}

	var getVars = func(vars []*pathVariable, paths []string) []pathFieldVar {
		var vv = make([]pathFieldVar, 0, len(vars))
		for _, v := range vars {
			pathVar := pathFieldVar{Fields: v.Fields}
			if v.end > 0 {
				pathVar.Value = strings.Join(paths[v.start:v.end+1], "/")
			} else {
				pathVar.Value = strings.Join(paths[v.start:], "/")
			}

			vv = append(vv, pathVar)
		}
		return vv
	}
	var getPath = func(nodes map[string]*pathNode, names ...string) *pathNode {
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
			return nil, errors.Wrapf(ErrPathNodeNotFound, "node=%s", n)
		}

		if vv := path.verbs[verb]; vv != nil && vv.Operation != nil && vv.Method == method {
			return &MatchOperation{
				Operation: generic.FromPtr(vv.Operation),
				Verb:      verb,
				Vars:      getVars(vv.Vars, paths),
			}, nil
		}
		nodes = path.nodes
	}

	return nil, errors.Wrapf(ErrNotFound, "method=%s path=%s", method, url)
}

type routeTarget struct {
	Method    string
	Operation *string
	Verb      *string
	Vars      []*pathVariable
}

type pathNode struct {
	nodes map[string]*pathNode
	verbs map[string]*routeTarget
}

type MatchOperation struct {
	Operation string
	Verb      string
	Vars      []pathFieldVar
}
