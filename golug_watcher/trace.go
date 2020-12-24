package golug_watcher

import (
	"github.com/pubgo/dix/dix_trace"
	"github.com/pubgo/xerror/xerror_util"
)

func init() {
	dix_trace.With(func(ctx *dix_trace.TraceCtx) {
		ctx.Func(Name+"_watch_callback", func() interface{} {
			var dt []string
			dataCallback.Range(func(key, _ interface{}) bool { dt = append(dt, key.(string)); return true })
			return dt
		})
		ctx.Func(Name+"_watcher", func() interface{} {
			var dt = make(map[string]string)
			registerMap.Range(func(key, val interface{}) bool {
				dt[key.(string)] = xerror_util.CallerWithFunc(val)
				return true
			})
			return dt
		})
	})
}
