package gateway

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"

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
