package lug

import (
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"
	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"go.uber.org/automaxprocs/maxprocs"

	_ "github.com/pubgo/lug/encoding/json"
	_ "github.com/pubgo/lug/encoding/protobuf"
	_ "github.com/pubgo/lug/healthy"
	_ "github.com/pubgo/lug/internal/debug"
	_ "github.com/pubgo/lug/logger"
	_ "github.com/pubgo/lug/metric"
	_ "github.com/pubgo/lug/registry/mdns"
	_ "github.com/pubgo/lug/tracing"
	"github.com/pubgo/lug/vars"
)

func init() {
	BeforeStart(func() {
		var log = maxprocs.Logger(func(s string, i ...interface{}) { xlog.Infof(s, i...) })
		var fn = xerror.ExitErr(maxprocs.Set(log))
		fn.(func())()
	})

	vars.Watch("dix", func() interface{} { return dix.Json() })
}
