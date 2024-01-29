package gateway

import (
	"bytes"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/log"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

func handlerWrap(path *httpPathRule) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var values = make(url.Values)
		for k, v := range path.vars {
			values.Set(k, v)
		}

		for k, v := range ctx.Queries() {
			values.Set(k, v)
		}

		path.opts.handlers[path.grpcMethodName].handler(path.opts, nil)

		contentType := ctx.Request().Header.ContentType()
		ctx.Protocol()
		ctx.Request().Header.Protocol()
		if r.ProtoMajor == 2 &&
			bytes.HasPrefix(contentType, []byte("application/grpc")) {
			m.serveGRPC(w, r)
			return
		}

		if strings.HasPrefix(contentType, "application/grpc-web") {
			m.serveGRPCWeb(w, r)
			return
		}

		r.URL.Path = "/" + strings.Trim(strings.TrimSpace(r.URL.Path), "/")
		if err := m.serveHTTP(w, r); err != nil {
			m.encError(w, r, err)
		}
	}
}

func fieldPath(fields protoreflect.FieldDescriptors, names ...string) []protoreflect.FieldDescriptor {
	fds := make([]protoreflect.FieldDescriptor, len(names))
	for i, name := range names {
		fd := fields.ByJSONName(name)
		if fd == nil {
			fd = fields.ByName(protoreflect.Name(name))
		}
		if fd == nil {
			return nil
		}

		fds[i] = fd

		// advance
		if i != len(fds)-1 {
			msgDesc := fd.Message()
			if msgDesc == nil {
				return nil
			}
			fields = msgDesc.Fields()
		}
	}
	return fds
}

func isGroup(path string) bool {
	re := regexp.MustCompile(`\{([^}]*)\}`)
	for _, match := range re.FindAllStringSubmatch(path, -1) {
		vars := strings.SplitN(match[1], "=", 2)
		if len(vars) == 2 {
			if strings.Contains(vars[1], "*") {
				return true
			}
		}
	}
	return false
}

func getPathVarMap(path string) map[string]string {
	var ret = make(map[string]string)
	re := regexp.MustCompile(`\{([^}]*)\}`)
	for _, match := range re.FindAllStringSubmatch(path, -1) {
		vars := strings.SplitN(match[1], "=", 2)
		field := vars[0]
		varName := strings.ReplaceAll(field, ".", "_")
		assert.If(strings.Contains(field, "*"), "field should not contain *, path=%s", path)
		if len(vars) == 2 {
			if strings.Contains(vars[1], "*") {
				ret[field] = "*"
			} else {
				log.Fatal().Str("name", vars[0]).Str("path", path).Msg("var field should contain *")
			}
		} else {
			ret[field] = varName
		}
	}
	return ret
}

// getPathVariables
//
// {field1}/field2/{field3}, {field1.abc}/field2/{field3.abc}
// {name=*}, {name=**}, {name.abc=*}
func getPathVariables(fields protoreflect.FieldDescriptors, path string) map[string][]protoreflect.FieldDescriptor {
	var ret = make(map[string][]protoreflect.FieldDescriptor)
	re := regexp.MustCompile(`\{([^}]*)\}`)
	for _, match := range re.FindAllStringSubmatch(path, -1) {
		vars := strings.SplitN(match[1], "=", 2)
		field := vars[0]
		varName := strings.ReplaceAll(field, ".", "_")
		assert.If(strings.Contains(field, "*"), "field should not contain *, path=%s", path)
		if len(vars) == 2 {
			if strings.Contains(vars[1], "*") {
				ret[varName] = fieldPath(fields, strings.Split(field, ".")...)
			} else {
				log.Fatal().Str("name", vars[0]).Str("path", path).Msg("var field should contain *")
			}
		} else {
			ret[varName] = fieldPath(fields, strings.Split(field, ".")...)
		}
	}
	return ret
}

func getMethod(rule *annotations.HttpRule, desc protoreflect.MethodDescriptor, name string) []*httpPathRule {
	var data []*httpPathRule

	var pathUrl, verb string
	switch v := rule.Pattern.(type) {
	case *annotations.HttpRule_Get:
		verb = http.MethodGet
		pathUrl = v.Get
	case *annotations.HttpRule_Put:
		verb = http.MethodPut
		pathUrl = v.Put
	case *annotations.HttpRule_Post:
		verb = http.MethodPost
		pathUrl = v.Post
	case *annotations.HttpRule_Delete:
		verb = http.MethodDelete
		pathUrl = v.Delete
	case *annotations.HttpRule_Patch:
		verb = http.MethodPatch
		pathUrl = v.Patch
	case *annotations.HttpRule_Custom:
		verb = strings.ToUpper(v.Custom.Kind)
		pathUrl = v.Custom.Path
	default:
		panic(fmt.Errorf("unsupported pattern %v", v))
	}

	assert.If(strings.Contains(pathUrl, ":"), "url should not contain ':'")

	inputFieldDescriptors := desc.Input().Fields()
	outputFieldDescriptors := desc.Output().Fields()

	// Method already registered.
	m := &httpPathRule{
		desc: desc,
		//vars:           getPathVariables(inputFieldDescriptors, pathUrl),
		vars:           getPathVarMap(pathUrl),
		grpcMethodName: name,
		rawHttpPath:    pathUrl,
		httpPath:       strings.ReplaceAll(pathUrl, ".", "_"),
		httpMethod:     verb,
		isGroup:        isGroup(pathUrl),
	}

	switch rule.Body {
	case "*":
		m.hasReqBody = true
	case "":
		m.hasReqBody = false
	default:
		m.hasReqBody = true
		m.reqBody = fieldPath(inputFieldDescriptors, strings.Split(rule.Body, ".")...)
	}

	switch rule.ResponseBody {
	case "":
		m.hasRspBody = false
	default:
		m.hasRspBody = true
		m.rspBody = fieldPath(outputFieldDescriptors, strings.Split(rule.ResponseBody, ".")...)
	}

	data = append(data, m)
	for _, addRule := range rule.AdditionalBindings {
		assert.If(len(addRule.AdditionalBindings) != 0, "nested rules")
		data = append(data, getMethod(addRule, desc, name)...)
	}
	return data
}

func quote(raw []byte) []byte {
	if n := len(raw); n > 0 && (raw[0] != '"' || raw[n-1] != '"') {
		raw = strconv.AppendQuote(raw[:0], string(raw))
	}
	return raw
}

// getExtensionHTTP
func getExtensionHTTP(m proto.Message) *annotations.HttpRule {
	return proto.GetExtension(m, annotations.E_Http).(*annotations.HttpRule)
}
