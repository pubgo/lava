package tracing

import (
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/strutil"
	"github.com/pubgo/funk/typex"
	"github.com/pubgo/lava/core/config"
)

var factories typex.SyncMap

// GetSpanID 从SpanContext中获取spanID
var GetSpanID = func(ctx opentracing.SpanContext) (string, string) { return "", "" }

type Factory func(cfg config.CfgMap) error

func GetFactory(names ...string) Factory {
	val, ok := factories.Load(strutil.GetDefault(names...))
	if !ok {
		return nil
	}

	return val.(Factory)
}

func RegisterFactory(name string, r Factory) {
	defer recovery.Exit()
	assert.Assert(name == "" || r == nil, "[name,tracer] is null")
	assert.Assert(factories.Has(name), "tracer %s already exists", name)
	factories.Set(name, r)
}
