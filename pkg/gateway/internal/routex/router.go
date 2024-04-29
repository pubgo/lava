// Copyright 2023-2024 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package routex

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func NewRouteTrie() *RouteTrie {
	return new(RouteTrie)
}

// RouteTrie is a prefix trie of valid REST URI paths to route targets.
// It supports evaluation of variables as the HttpPath is matched, for
// interpolating parts of the URI HttpPath into an RPC request field. The
// map is keyed by the HttpPath component that corresponds to a given node.
type RouteTrie struct {
	// Child nodes, keyed by the next segment in the HttpPath.
	children map[string]*RouteTrie
	// Final node in the HttpPath has a map of verbs to methods.
	// Verbs are either an empty string or a single literal.
	verbs map[string]RouteMethods
}

func (t *RouteTrie) GetRouteMethods() []*RouteTarget {
	var rr []*RouteTarget
	for _, verbs := range t.verbs {
		for _, target := range verbs {
			rr = append(rr, target)
		}
	}

	for _, child := range t.children {
		rr = append(rr, child.GetRouteMethods()...)
	}

	return rr
}

func (t *RouteTrie) Match(uriPath, httpMethod string) (*RouteTarget, []RouteTargetVarMatch, RouteMethods) {
	return t.match(uriPath, httpMethod)
}

func (t *RouteTrie) AddRoute(grpcMethod string, mth protoreflect.MethodDescriptor) error {
	httpRule := getExtensionHTTP(mth)
	return t.addRoute(&MethodConfig{Descriptor: mth, MethodPath: grpcMethod}, httpRule)
}

// addRoute adds a target to the router for the given HttpMethod and the given
// HTTP rule. Only the rule itself is added. If the rule indicates additional
// bindings, they are ignored. To add routes for all bindings, callers must
// invoke this HttpMethod for each rule.
func (t *RouteTrie) addRoute(config *MethodConfig, httpRule *annotations.HttpRule) error {
	if httpRule == nil {
		return nil
	}

	var method, template string
	switch pattern := httpRule.GetPattern().(type) {
	case *annotations.HttpRule_Get:
		method, template = http.MethodGet, pattern.Get
	case *annotations.HttpRule_Put:
		method, template = http.MethodPut, pattern.Put
	case *annotations.HttpRule_Post:
		method, template = http.MethodPost, pattern.Post
	case *annotations.HttpRule_Delete:
		method, template = http.MethodDelete, pattern.Delete
	case *annotations.HttpRule_Patch:
		method, template = http.MethodPatch, pattern.Patch
	case *annotations.HttpRule_Custom:
		method, template = pattern.Custom.GetKind(), pattern.Custom.GetPath()
	default:
		return fmt.Errorf("invalid type of pattern for HTTP httpRule: %T", pattern)
	}

	if method == "" {
		return errors.New("invalid HTTP httpRule: HttpMethod is blank")
	}
	if template == "" {
		return errors.New("invalid HTTP httpRule: HttpPath template is blank")
	}
	segments, variables, err := parsePathTemplate(template)
	if err != nil {
		return err
	}

	target, err := makeTarget(config, method, httpRule.GetBody(), httpRule.GetResponseBody(), segments, variables)
	if err != nil {
		return err
	}
	if err := t.insert(method, target, segments); err != nil {
		return err
	}

	for i, rule := range httpRule.GetAdditionalBindings() {
		if len(rule.GetAdditionalBindings()) > 0 {
			return fmt.Errorf("nested additional bindings are not supported (HttpMethod %s)", config.MethodPath)
		}
		if err = t.addRoute(config, rule); err != nil {
			return fmt.Errorf("failed to add REST route (add'l binding #%d) for HttpMethod %s: %w", i+1, config.MethodPath, err)
		}
	}

	return nil
}

func (t *RouteTrie) insertChild(segment string) *RouteTrie {
	child := t.children[segment]
	if child == nil {
		if t.children == nil {
			t.children = make(map[string]*RouteTrie, 1)
		}
		child = &RouteTrie{}
		t.children[segment] = child
	}
	return child
}

func (t *RouteTrie) insertVerb(verb string) RouteMethods {
	methods := t.verbs[verb]
	if methods == nil {
		if t.verbs == nil {
			t.verbs = make(map[string]RouteMethods, 1)
		}
		methods = make(RouteMethods, 1)
		t.verbs[verb] = methods
	}
	return methods
}

// insert the target into the trie using the given HttpMethod and segment HttpPath.
// The HttpPath is followed until the final segment is reached.
func (t *RouteTrie) insert(method string, target *RouteTarget, segments pathSegments) error {
	cursor := t
	for _, segment := range segments.path {
		cursor = cursor.insertChild(segment)
	}
	if existing := cursor.verbs[segments.verb][method]; existing != nil {
		return alreadyExistsError{
			existing:    existing,
			pathPattern: segments.String(),
			method:      method,
		}
	}
	cursor.insertVerb(segments.verb)[method] = target
	return nil
}

// match finds a route for the given request. If a match is found, the associated target and a map
// of matched variable values is returned.
func (t *RouteTrie) match(uriPath, httpMethod string) (*RouteTarget, []RouteTargetVarMatch, RouteMethods) {
	if len(uriPath) == 0 || uriPath[0] != '/' || uriPath[len(uriPath)-1] == ':' {
		// Must start with "/" or if it ends with ":" it won't match
		return nil, nil, nil
	}
	uriPath = uriPath[1:] // skip the leading slash

	path := strings.Split(uriPath, "/")
	var verb string
	if len(path) > 0 {
		lastElement := path[len(path)-1]
		if pos := strings.IndexRune(lastElement, ':'); pos >= 0 {
			path[len(path)-1] = lastElement[:pos]
			verb = lastElement[pos+1:]
		}
	}
	target, methods := t.findTarget(path, verb, httpMethod)
	if target == nil {
		return nil, nil, methods
	}

	vars, err := computeVarValues(path, target)
	if err != nil {
		log.Err(err).Msg("failed to compute var values")
		return nil, nil, nil
	}

	return target, vars, nil
}

// findTarget finds the target for the given HttpPath components, Verb, and HttpMethod.
// The HttpMethod either returns a target OR the set of methods for the given HttpPath
// and Verb. If the target is non-nil, the request was matched. If the target
// is nil but methods are non-nil, the HttpPath and Verb matched a route, but not
// the HttpMethod. This can be used to send back a well-formed "Allow" response
// header. If both are nil, the HttpPath and Verb did not match.
func (t *RouteTrie) findTarget(path []string, verb, method string) (*RouteTarget, RouteMethods) {
	if len(path) == 0 {
		return t.getTarget(verb, method)
	}
	current := path[0]
	path = path[1:]

	if child := t.children[current]; child != nil {
		target, methods := child.findTarget(path, verb, method)
		if target != nil || methods != nil {
			return target, methods
		}
	}

	if childAst := t.children["*"]; childAst != nil {
		target, methods := childAst.findTarget(path, verb, method)
		if target != nil || methods != nil {
			return target, methods
		}
	}

	// Double-asterisk must be the last element in pattern.
	// So it consumes all remaining HttpPath elements.
	if childDblAst := t.children["**"]; childDblAst != nil {
		return childDblAst.findTarget(nil, verb, method)
	}
	return nil, nil
}

// getTarget gets the target for the given Verb and HttpMethod from the
// node trie. It is like findTarget, except that it does not use a
// HttpPath to first descend into a sub-trie.
func (t *RouteTrie) getTarget(verb, method string) (*RouteTarget, RouteMethods) {
	methods := t.verbs[verb]
	if target := methods[method]; target != nil {
		return target, methods
	}
	// See if a wildcard HttpMethod was used
	if target := methods["*"]; target != nil {
		return target, methods
	}
	return nil, methods
}

type RouteMethods map[string]*RouteTarget

type RouteTarget struct {
	GrpcMethodName string
	HttpMethod     string
	HttpPath       []string
	Verb           string

	RequestBodyFieldPath  string
	RequestBodyFields     []protoreflect.FieldDescriptor
	ResponseBodyFieldPath string
	ResponseBodyFields    []protoreflect.FieldDescriptor
	Vars                  []RouteTargetVar
}

func makeTarget(config *MethodConfig, method, requestBody, responseBody string, segments pathSegments, variables []pathVariable) (*RouteTarget, error) {
	requestBodyFields, err := resolvePathToDescriptors(config.Descriptor.Input(), requestBody)
	if err != nil {
		return nil, err
	}
	if len(requestBodyFields) > 1 {
		return nil, fmt.Errorf(
			"unexpected request body HttpPath %q: must be a single field",
			requestBody,
		)
	}

	responseBodyFields, err := resolvePathToDescriptors(config.Descriptor.Output(), responseBody)
	if err != nil {
		return nil, err
	}
	if len(responseBodyFields) > 1 {
		return nil, fmt.Errorf(
			"unexpected response body HttpPath %q: must be a single field",
			requestBody,
		)
	}

	routeTargetVars := make([]RouteTargetVar, len(variables))
	for i, variable := range variables {
		fields, err := resolvePathToDescriptors(config.Descriptor.Input(), variable.FieldPath)
		if err != nil {
			return nil, err
		}
		if last := fields[len(fields)-1]; last.IsList() {
			return nil, fmt.Errorf(
				"unexpected HttpPath variable %q: cannot be a repeated field",
				variable.FieldPath,
			)
		}

		routeTargetVars[i] = RouteTargetVar{
			pathVariable: variable,
			fields:       fields,
		}
	}
	return &RouteTarget{
		GrpcMethodName:        config.MethodPath,
		HttpMethod:            method,
		HttpPath:              segments.path,
		Verb:                  segments.verb,
		RequestBodyFieldPath:  requestBody,
		RequestBodyFields:     requestBodyFields,
		ResponseBodyFieldPath: responseBody,
		ResponseBodyFields:    responseBodyFields,
		Vars:                  routeTargetVars,
	}, nil
}

type RouteTargetVar struct {
	pathVariable

	fields []protoreflect.FieldDescriptor
}

func (v RouteTargetVar) size() int {
	if v.end == -1 {
		return -1
	}
	return v.end - v.start
}

func (v RouteTargetVar) index(segments []string) []string {
	start, end := v.start, v.end
	if end == -1 {
		if start >= len(segments) {
			return nil
		}
		return segments[start:]
	}
	return segments[start:end]
}

func (v RouteTargetVar) capture(segments []string) (string, error) {
	parts := v.index(segments)
	mode := pathEncodeSingle
	if v.end == -1 || v.start-v.end > 1 {
		mode = pathEncodeMulti
	}

	var sb strings.Builder
	for i, part := range parts {
		val, err := pathUnescape(part, mode)
		if err != nil {
			return "", err
		}
		if i > 0 {
			sb.WriteByte('/')
		}
		sb.WriteString(val)
	}
	return sb.String(), nil
}

type RouteTargetVarMatch struct {
	Fields []protoreflect.FieldDescriptor
	Value  string
	Name   string
}

func computeVarValues(path []string, target *RouteTarget) ([]RouteTargetVarMatch, error) {
	if len(target.Vars) == 0 {
		return nil, nil
	}
	vars := make([]RouteTargetVarMatch, len(target.Vars))
	for i, varDef := range target.Vars {
		val, err := varDef.capture(path)
		if err != nil {
			return nil, err
		}
		vars[i].Fields = varDef.fields
		vars[i].Value = val
		vars[i].Name = varDef.FieldPath
	}
	return vars, nil
}

// resolvePathToDescriptors translates the given HttpPath string, in the form of "ident.ident.ident",
// into a HttpPath of FieldDescriptors, relative to the given msg.
func resolvePathToDescriptors(msg protoreflect.MessageDescriptor, path string) ([]protoreflect.FieldDescriptor, error) {
	if path == "" {
		return nil, nil
	}
	if path == "*" {
		// non-nil, empty slice means use the whole thing
		return []protoreflect.FieldDescriptor{}, nil
	}

	fields := msg.Fields()
	parts := strings.Split(path, ".")
	result := make([]protoreflect.FieldDescriptor, len(parts))
	for i, part := range parts {
		field := fields.ByName(protoreflect.Name(part))
		if field == nil {
			return nil, fmt.Errorf("in field HttpPath %q: element %q does not correspond to any field of type %s",
				path, part, msg.FullName())
		}
		result[i] = field
		if i == len(parts)-1 {
			break
		}
		if field.Cardinality() == protoreflect.Repeated {
			return nil, fmt.Errorf("in field HttpPath %q: field %q of type %s should not be a list or map",
				path, part, msg.FullName())
		}
		msg = field.Message()
		if msg == nil {
			return nil, fmt.Errorf("in field HttpPath %q: field %q of type %s should be a message but is instead %s",
				path, part, msg.FullName(), field.Kind())
		}
		fields = msg.Fields()
	}
	return result, nil
}

// resolveFieldDescriptorsToPath translates the given HttpPath of FieldDescriptors into a string
// of the form "ident.ident.ident".
func resolveFieldDescriptorsToPath(fields []protoreflect.FieldDescriptor) string {
	if len(fields) == 0 {
		return ""
	}
	parts := make([]string, len(fields))
	for i, field := range fields {
		parts[i] = string(field.Name())
	}
	return strings.Join(parts, ".")
}

type alreadyExistsError struct {
	existing            *RouteTarget
	pathPattern, method string
}

func (a alreadyExistsError) Error() string {
	return fmt.Sprintf("target for %s, HttpMethod %s already exists: %s", a.pathPattern, a.method, a.existing.GrpcMethodName)
}

// getExtensionHTTP
func getExtensionHTTP(m protoreflect.MethodDescriptor) *annotations.HttpRule {
	if m == nil {
		return nil
	}

	if m.Options() == nil {
		return nil
	}

	ext, ok := proto.GetExtension(m.Options(), annotations.E_Http).(*annotations.HttpRule)
	if ok {
		return ext
	}
	return nil
}
