package golug_watcher

import (
	"github.com/pubgo/dix/dix_trace"
	"github.com/pubgo/xprocess/xutil"
)

func init() {
	dix_trace.With(func(ctx *dix_trace.Ctx) {
		ctx.Func(Name+"_watcher_callback", func() interface{} {
			var dt []string
			callbackMap.Each(func(key string) { dt = append(dt, key) })
			return dt
		})

		ctx.Func(Name+"_watcher", func() interface{} {
			var dt = make(map[string]string)
			for k, v := range List() {
				dt[k] = xutil.FuncStack(v)
			}
			return dt
		})
	})
}
