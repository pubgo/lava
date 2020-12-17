package golug_plugin

import (
	"expvar"

	"github.com/pubgo/dix/dix_trace"
)

func init() {
	dix_trace.With(func(_ *dix_trace.TraceCtx) {
		expvar.Publish("plugin", expvar.Func(func() interface{} { return All() }))
	})
}
