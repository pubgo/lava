package tracer

import (
	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/types"
	"github.com/pubgo/xerror"
)

type Factory func(cfg map[string]interface{}) (opentracing.Tracer, error)

var tracers types.SMap

func Get(names ...string) Factory {
	val, ok := tracers.Load(consts.GetDefault(names...))
	if !ok {
		return nil
	}

	return val.(Factory)
}

func Register(name string, r Factory) (err error) {
	defer xerror.RespErr(&err)

	xerror.Assert(name == "" || r == nil, "[name,tracer] is null")
	xerror.Assert(tracers.Has(name), "tracer %s already exists", name)
	tracers.Set(name, r)
	return
}
