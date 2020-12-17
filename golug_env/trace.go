package golug_env

import (
	"expvar"
	"os"

	"github.com/pubgo/dix/dix_trace"
)

func init() {
	dix_trace.With(func(_ *dix_trace.TraceCtx) {
		expvar.Publish("envs", expvar.Func(func() interface{} { return os.Environ() }))
	})
}
