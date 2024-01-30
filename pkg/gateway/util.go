package gateway

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/lava/pkg/wsutil"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"io"
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
			values.Set(k, v)
		}

		for k, v := range ctx.Queries() {
			values.Set(k, v)
		}

		var h = path.opts.handlers[path.grpcMethodName]

		if wsutil.IsWebSocketUpgrade(ctx) {
			if !h.desc.IsStreamingClient() || !h.desc.IsStreamingServer() {
				return errors.Format("服务不支持 websocket")
			}

			conn, err := wsutil.New(ctx)
			err := h.handler(path.opts, &streamWS{
				ctx:      ctx.Context(),
				conn:     conn,
				pathRule: path,
			})
			return nil
		}

		err := h.handler(path.opts, &streamHTTP{
			ctx:    ctx.Context(),
			method: method,
			params: params,
			opts:   m.opts,

			// write
			w:       ctx,
			wHeader: w.Header(),

			// read
			r:       ctx.Body(),
			rHeader: r.Header,

			contentType:    contentType,
			accept:         accept,
			acceptEncoding: acceptEncoding,
			hasBody:        r.ContentLength > 0 || r.ContentLength == -1,
		})
		return err
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

	m.hasRspBody = true
	if rule.ResponseBody != "" {
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

// AsHTTPBodyWriter returns the writer of a stream of google.api.HttpBody.
// The first message will be marshalled from msg excluding the data field.
// The returned writer is only valid during the lifetime of the RPC.
func AsHTTPBodyWriter(stream grpc.ServerStream, msg proto.Message) (body io.Writer, err error) {
	ctx := stream.Context()
	s, err := streamHTTPFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if !s.method.desc.IsStreamingServer() {
		return nil, fmt.Errorf("expected streaming server")
	}
	if s.sendCount > 0 {
		return nil, fmt.Errorf("expected first message")
	}

	cur := msg.ProtoReflect()
	if name, want := cur.Descriptor().FullName(), s.method.desc.Output().FullName(); name != want {
		return nil, fmt.Errorf("expected %s got %s", want, name)
	}
	for _, fd := range s.method.rspBody {
		cur = cur.Mutable(fd).Message()
	}

	if typ := cur.Descriptor().FullName(); typ != "google.api.HttpBody" {
		return nil, fmt.Errorf("expected body type of google.api.HttpBody got %s", typ)
	}

	fds := cur.Descriptor().Fields()
	fdContentType := fds.ByName("content_type")
	pContentType := cur.Get(fdContentType)
	contentType := pContentType.String()

	s.wHeader.Set("Content-Type", contentType)
	if !s.sentHeader {
		if err := s.SendHeader(nil); err != nil {
			return nil, err
		}
	}
	s.sendCount += 1
	return s.w, nil
}

func streamHTTPFromCtx(ctx context.Context) (*streamHTTP, error) {
	ss := grpc.ServerTransportStreamFromContext(ctx)
	if ss == nil {
		return nil, fmt.Errorf("invalid server transport stream")
	}
	sts, ok := ss.(*serverTransportStream)
	if !ok {
		return nil, fmt.Errorf("unknown server transport stream")
	}
	s, ok := sts.ServerStream.(*streamHTTP)
	if !ok {
		return nil, fmt.Errorf("expected HTTP stream got %T", sts.ServerStream)
	}
	return s, nil
}

// AsHTTPBodyReader returns the reader of a stream of google.api.HttpBody.
// The first message will be unmarshalled into msg excluding the data field.
// The returned reader is only valid during the lifetime of the RPC.
func AsHTTPBodyReader(stream grpc.ServerStream, msg proto.Message) (body io.Reader, err error) {
	ctx := stream.Context()
	s, err := streamHTTPFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if !s.method.desc.IsStreamingClient() {
		return nil, fmt.Errorf("expected streaming client")
	}
	if s.recvCount > 0 {
		return nil, fmt.Errorf("expected first message")
	}

	cur := msg.ProtoReflect()
	if name, want := cur.Descriptor().FullName(), s.method.desc.Input().FullName(); name != want {
		return nil, fmt.Errorf("expected %s got %s", want, name)
	}
	for _, fd := range s.method.reqBody {
		cur = cur.Mutable(fd).Message()
	}

	if typ := cur.Descriptor().FullName(); typ != "google.api.HttpBody" {
		return nil, fmt.Errorf("expected body type of google.api.HttpBody got %s", typ)
	}

	fds := cur.Descriptor().Fields()
	fdContentType := fds.ByName("content_type")
	cur.Set(fdContentType, protoreflect.ValueOfString(s.contentType))
	// TODO: extensions?

	if err := s.params.set(msg); err != nil {
		return nil, err
	}
	s.recvCount += 1
	return s.r, nil
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
