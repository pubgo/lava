package golug_plugin

import (
	"github.com/pubgo/dix/dix_trace"
)

func init() {
	dix_trace.With(func(ctx *dix_trace.TraceCtx) {
		ctx.Func("plugin", func() interface{} {
			var data = make(map[string][]string)
			for k, v := range All() {
				for i := range v {
					data[k] = append(data[k], v[i].String())
				}
			}
			return data
		})
	})
}
