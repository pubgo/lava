package golug_config

import (
	"expvar"

	"github.com/pubgo/dix/dix_trace"
)

func init() {
	dix_trace.With(func(_ *dix_trace.TraceCtx) {
		expvar.Publish(Name, expvar.Func(func() interface{} { return GetCfg().AllSettings() }))
	})
}
