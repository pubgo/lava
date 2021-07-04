package lug

import (
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"
	"github.com/pubgo/dix"
	_ "github.com/pubgo/lug/encoding/json"
	_ "github.com/pubgo/lug/encoding/protobuf"
	_ "github.com/pubgo/lug/healthy"
	_ "github.com/pubgo/lug/internal/debug"
	_ "github.com/pubgo/lug/internal/log"
	_ "github.com/pubgo/lug/metric"
	_ "github.com/pubgo/lug/tracing"
	"github.com/pubgo/lug/vars"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"go.uber.org/automaxprocs/maxprocs"
)

func init() {
	var fn = xerror.ExitErr(maxprocs.Set(maxprocs.Logger(func(s string, i ...interface{}) { xlog.Infof(s, i...) })))
	fn.(func())()

	vars.Watch("dix", func() interface{} { return dix.Json() })
}
