package tracing

import (
	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/lava/config/config_type"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/pkg/utils"
)

var factories typex.SMap

// GetSpanID 从SpanContext中获取spanID
var GetSpanID = func(ctx opentracing.SpanContext) (string, string) { return "", "" }

type Factory func(cfg config_type.CfgMap) error

func GetFactory(names ...string) Factory {
	val, ok := factories.Load(utils.GetDefault(names...))
	if !ok {
		return nil
	}

	return val.(Factory)
}

func RegisterFactory(name string, r Factory) {
	defer xerror.RespExit(name)
	xerror.Assert(name == "" || r == nil, "[name,tracer] is null")
	xerror.Assert(factories.Has(name), "tracer %s already exists", name)
	factories.Set(name, r)
}
