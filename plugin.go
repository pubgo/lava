package lug

import (
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"
	"github.com/pubgo/dix"
	_ "github.com/pubgo/lug/debug"
	_ "github.com/pubgo/lug/encoding/json"
	_ "github.com/pubgo/lug/encoding/protobuf"
	_ "github.com/pubgo/lug/healthy"
	_ "github.com/pubgo/lug/internal/log"
	_ "github.com/pubgo/lug/metric"
	_ "github.com/pubgo/lug/tracing"
	"github.com/pubgo/lug/vars"
	_ "go.uber.org/automaxprocs"
)

func init() {
	vars.Watch("dix", func() interface{} { return dix.Json() })
}
