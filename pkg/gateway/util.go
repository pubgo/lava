package gateway

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"

	"github.com/pubgo/funk/errors"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

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

func quote(raw []byte) []byte {
	if n := len(raw); n > 0 && (raw[0] != '"' || raw[n-1] != '"') {
		raw = strconv.AppendQuote(raw[:0], string(raw))
	}
	return raw
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

	if err := cb(method, template, httpRule.GetBody(), httpRule.GetResponseBody()); err != nil {
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
