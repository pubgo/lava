package golug_broker

import (
	"github.com/pubgo/dix/dix_trace"
)

func init() {
	dix_trace.With(func(ctx *dix_trace.TraceCtx) {
		ctx.Func(Name, func() interface{} {
			var data = make(map[string]string)
			for k, v := range List() {
				data[k] = v.Name()
			}
			return data
		})
	})
}
