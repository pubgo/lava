package golug_watcher

import (
	"github.com/pubgo/dix/dix_trace"
)

func init() {
	dix_trace.With(func(ctx *dix_trace.TraceCtx) {
		ctx.Func(Name, func() interface{} {
			var dt []string
			dataCallback.Range(func(key, _ interface{}) bool { dt = append(dt, key.(string)); return true })
			return dt
		})
	})
}
