package traceRecord

import (
	"encoding/base64"
	"strings"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/metadata"
)

const (
	binHdrSuffix = "-bin"
)

var _ opentracing.TextMapWriter = (textMapCarrier)(nil)
var _ opentracing.TextMapReader = (textMapCarrier)(nil)

// textMapCarrier extends a metadata.MD to be an opentracing textMap
type textMapCarrier metadata.MD

// Set is a opentracing.TextMapReader interface that extracts values.
func (m textMapCarrier) Set(key, val string) {
	// gRPC allows for complex binary values to be written.
	encodedKey, encodedVal := encodeKeyValue(key, val)
	// The metadata object is a multimap, and previous values may exist, but for opentracing headers, we do not append
	// we just override.
	m[encodedKey] = []string{encodedVal}
}

// ForeachKey is a opentracing.TextMapReader interface that extracts values.
func (m textMapCarrier) ForeachKey(callback func(key, val string) error) error {
	for k, vv := range m {
		for _, v := range vv {
			if err := callback(k, v); err != nil {
				return err
			}
		}
	}
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
