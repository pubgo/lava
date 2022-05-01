package tracing_middleware

import (
	"encoding/base64"
	"github.com/pubgo/lava/middleware"
	"github.com/pubgo/xerror"
	"strings"

	"github.com/opentracing/opentracing-go"
)

const (
	binHdrSuffix = "-bin"
)

var _ opentracing.TextMapWriter = (*textMapCarrier)(nil)
var _ opentracing.TextMapReader = (*textMapCarrier)(nil)

// textMapCarrier extends a metadata.MD to be an opentracing textMap
type textMapCarrier struct {
	*middleware.RequestHeader
}

// Set is a opentracing.TextMapReader interface that extracts values.
func (m *textMapCarrier) Set(key, val string) {
	// gRPC allows for complex binary values to be written.
	encodedKey, encodedVal := encodeKeyValue(key, val)
	// The metadata object is a multimap, and previous values may exist, but for opentracing headers, we do not append
	// we just override.
	m.RequestHeader.Set(encodedKey, encodedVal)
}

// ForeachKey is a opentracing.TextMapReader interface that extracts values.
func (m *textMapCarrier) ForeachKey(callback func(key, val string) error) error {
	m.VisitAll(func(key, value []byte) {
		xerror.Panic(callback(string(key), string(value)))
	})
	return nil
}

// encodeKeyValue encodes key and value qualified for transmission via gRPC.
// note: copy pasted from private values of grpc.metadata
func encodeKeyValue(k, v string) (string, string) {
	k = strings.ToLower(k)
	if strings.HasSuffix(k, binHdrSuffix) {
		val := base64.StdEncoding.EncodeToString([]byte(v))
		v = val
	}
	return k, v
}
