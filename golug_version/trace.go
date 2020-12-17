package golug_version

import (
	"expvar"

	"github.com/pubgo/dix/dix_trace"
)

func init() {
	dix_trace.With(func(_ *dix_trace.TraceCtx) {
		expvar.Publish("envs", expvar.Func(func() interface{} { return List() }))
	})
}
