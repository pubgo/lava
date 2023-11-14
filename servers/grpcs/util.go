package grpcs

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/v2/utilities"
	"google.golang.org/protobuf/proto"
	"net/url"
	"strings"
)

// serviceFromMethod returns the service
// /service.Foo/Bar => service.Foo
func serviceFromMethod(m string) string {
	if len(m) == 0 {
		return m
	}

	return strings.Split(strings.Trim(m, "/"), "/")[0]
}

type DefaultQueryParser struct{}

// Parse populates "values" into "msg".
// A value is ignored if its key starts with one of the elements in "filter".
func (*DefaultQueryParser) Parse(msg proto.Message, values url.Values, filter *utilities.DoubleArray) error {
	for key, v := range values {
		if len(v) == 0 {
			delete(values, key)
			continue
		}

		if len(v) == 1 && v[0] == "" {
			delete(values, key)
			continue
		}
	}

	return new(runtime.DefaultQueryParser).Parse(msg, values, filter)
}
