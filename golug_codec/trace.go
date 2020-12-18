package golug_codec

import (
	"expvar"

	"github.com/pubgo/dix/dix_trace"
)

func init() {
	dix_trace.With(func(_ *dix_trace.TraceCtx) {
		expvar.Publish(Name, expvar.Func(func() interface{} {
			var dt []string
			data.Range(func(key, value interface{}) bool {
				dt = append(dt, key.(string))
				return true
			})
			return dt
		}))
	})
}
