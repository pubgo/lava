package gateway

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/textproto"
	"strings"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/lava/pkg/gateway/internal/routertree"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func getReqBodyDesc(path *routertree.MatchOperation) []protoreflect.FieldDescriptor {
	return path.Extras["req_body_desc"].([]protoreflect.FieldDescriptor)
}

func getRspBodyDesc(path *routertree.MatchOperation) []protoreflect.FieldDescriptor {
	return path.Extras["rsp_body_desc"].([]protoreflect.FieldDescriptor)
}

func resolveBodyDesc(methodDesc protoreflect.MethodDescriptor, reqBody, rspBody string) map[string]any {
	return map[string]any{
		"req_body_field": reqBody,
		"req_body_desc":  assert.Must1(resolvePathToDescriptors(methodDesc.Input(), reqBody)),
		"rsp_body_field": rspBody,
		"rsp_body_desc":  assert.Must1(resolvePathToDescriptors(methodDesc.Output(), rspBody)),
	}
}

func fieldPathToDesc(fields protoreflect.FieldDescriptors, names ...string) []protoreflect.FieldDescriptor {
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

// getExtensionHTTP
func getExtensionHTTP(m protoreflect.MethodDescriptor) *annotations.HttpRule {
	if m == nil || m.Options() == nil {
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

func handlerHttpRoute(httpRule *annotations.HttpRule, cb func(mth string, path string, reqBody, rspBody string) error) error {
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
		return errors.Format("invalid type of pattern for HTTP httpRule: %T", pattern)
	}

	if method == "" {
		return errors.New("invalid HTTP httpRule: HttpMethod is blank")
	}

	if template == "" {
		return errors.New("invalid HTTP httpRule: HttpPath template is blank")
	}

	var reqBody = httpRule.GetBody()
	switch reqBody {
	case "", "*":
		reqBody = "*"
	}

	var rspBody = httpRule.GetResponseBody()
	switch rspBody {
	case "", "*":
		rspBody = "*"
	}

	if err := cb(method, template, reqBody, rspBody); err != nil {
		return err
	}

	for i, rule := range httpRule.GetAdditionalBindings() {
		if len(rule.GetAdditionalBindings()) > 0 {
			return errors.New("nested additional bindings are not supported")
		}

		if err := handlerHttpRoute(rule, cb); err != nil {
			return errors.Format("failed to add REST route (add'l binding #%d): %w", i+1, err)
		}
	}

	return nil
}

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
			return nil, errors.Format("in field HttpPath %q: element %q does not correspond to any field of type %s",
				path, part, msg.FullName())
		}

		result[i] = field
		if i == len(parts)-1 {
			break
		}

		if field.Cardinality() == protoreflect.Repeated {
			return nil, errors.Format("in field HttpPath %q: field %q of type %s should not be a list or map", path, part, msg.FullName())
		}

		msg = field.Message()
		if msg == nil {
			return nil, fmt.Errorf("in field HttpPath %q: field %q of type %s should be a message", path, part, field.Kind())
		}

		fields = msg.Fields()
	}
	return result, nil
}
