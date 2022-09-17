package tracing

import (
	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/pkg/utils"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
)

var factories typex.SMap

// GetSpanID 从SpanContext中获取spanID
var GetSpanID = func(ctx opentracing.SpanContext) (string, string) { return "", "" }

type Factory func(cfg config.CfgMap) error

func GetFactory(names ...string) Factory {
	val, ok := factories.Load(utils.GetDefault(names...))
	if !ok {
		return nil
	}

	return val.(Factory)
}

func RegisterFactory(name string, r Factory) {
	defer xerror.RecoverAndExit()
	xerror.Assert(name == "" || r == nil, "[name,tracer] is null")
	xerror.Assert(factories.Has(name), "tracer %s already exists", name)
	factories.Set(name, r)
}
