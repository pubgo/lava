package gateway

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/lava/pkg/wsutil"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"net/http"
	"net/textproto"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

func handlerWrap(path *httpPathRule) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var values = make(url.Values)
		for k, v := range path.vars {
			values.Set(k, ctx.Params(v))
		}

		for k, v := range ctx.Queries() {
			values.Set(k, v)
		}

		var doRequest = path.opts.handlers[path.grpcMethodName]

		fmt.Println("IsWebSocketUpgrade", wsutil.IsWebSocketUpgrade(ctx))
		if wsutil.IsWebSocketUpgrade(ctx) {
			if !path.desc.IsStreamingClient() || !path.desc.IsStreamingServer() {
				return errors.Format("服务不支持 websocket")
			}

			return wsutil.New(ctx, func(c *websocket.Conn) {
				fmt.Println(doRequest(&streamWS{
					ctx:      ctx,
					conn:     c,
					pathRule: path,
					params:   values,
				}))
			})
		}

		return doRequest(&streamHTTP{
			ctx:    ctx,
			method: path,
			params: values,
			opts:   path.opts,
		})
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

func getMethod(opts *muxOptions, rule *annotations.HttpRule, desc protoreflect.MethodDescriptor, name string) []*httpPathRule {
	if rule == nil {
		return nil
	}

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
		panic(fmt.Errorf("unsupported http rule pattern %v", v))
	}

	assert.If(strings.Contains(pathUrl, ":"), "grpc http rule pattern url should not contain ':'")

	// Method already registered.
	m := &httpPathRule{
		opts: opts,
		desc: desc,
		//vars:           getPathVariables(inputFieldDescriptors, pathUrl),
		grpcMethodName: name,
		rawHttpPath:    pathUrl,
		httpMethod:     verb,

		// TODO 未来需要调整
		vars:     getPathVarMap(pathUrl),
		httpPath: strings.ReplaceAll(pathUrl, ".", "_"),
		isGroup:  isGroup(pathUrl),
	}

	switch rule.Body {
	case "*":
		m.hasReqBody = true
	case "":
		m.hasReqBody = false
	default:
		m.hasReqBody = true
		inputFieldDescriptors := desc.Input().Fields()
		m.reqBody = fieldPath(inputFieldDescriptors, strings.Split(rule.Body, ".")...)
	}
	if verb == http.MethodGet {
		m.hasReqBody = false
	}

	m.hasRspBody = true
	if rule.ResponseBody != "" {
		outputFieldDescriptors := desc.Output().Fields()
		m.rspBody = fieldPath(outputFieldDescriptors, strings.Split(rule.ResponseBody, ".")...)
	}

	var data = []*httpPathRule{m}
	for _, addRule := range rule.AdditionalBindings {
		assert.If(len(addRule.AdditionalBindings) != 0, "nested rules are not allowed")
		for _, m1 := range getMethod(opts, addRule, desc, name) {
			if m1 == nil {
				continue
			}

			data = append(data, m1)
		}
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

func setOutgoingHeader(header http.Header, md metadata.MD) {
	for k, vs := range md {
		if isReservedHeader(k) {
			continue
		}

		if strings.HasSuffix(k, binHdrSuffix) {
			dst := make([]string, len(vs))
			for i, v := range vs {
				dst[i] = encodeBinHeader([]byte(v))
			}
			vs = dst
		}
		header[textproto.CanonicalMIMEHeaderKey(k)] = vs
	}
}

func isReservedHeader(k string) bool {
	switch k {
	case "content-type", "user-agent", "grpc-message-type", "grpc-encoding",
		"grpc-message", "grpc-status", "grpc-timeout",
		"grpc-status-details", "te":
		return true
	default:
		return false
	}
}
func isWhitelistedHeader(k string) bool {
	switch k {
	case ":authority", "user-agent":
		return true
	default:
		return false
	}
}

const binHdrSuffix = "-bin"

func encodeBinHeader(b []byte) string {
	return base64.RawStdEncoding.EncodeToString(b)
}

func decodeBinHeader(v string) (s string, err error) {
	var b []byte
	if len(v)%4 == 0 {
		// Input was padded, or padding was not necessary.
		b, err = base64.RawStdEncoding.DecodeString(v)
	} else {
		b, err = base64.RawStdEncoding.DecodeString(v)
	}
	return string(b), err
}

func newIncomingContext(ctx context.Context, header http.Header) (context.Context, metadata.MD) {
	md := make(metadata.MD, len(header))
	for k, vs := range header {
		k = strings.ToLower(k)
		if isReservedHeader(k) && !isWhitelistedHeader(k) {
			continue
		}
		if strings.HasSuffix(k, binHdrSuffix) {
			dst := make([]string, len(vs))
			for i, v := range vs {
				v, err := decodeBinHeader(v)
				if err != nil {
					continue // TODO: log error?
				}
				dst[i] = v
			}
			vs = dst
		}
		md[k] = vs
	}
	return metadata.NewIncomingContext(ctx, md), md
}
