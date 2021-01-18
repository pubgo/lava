package golug_codec

import (
	"github.com/pubgo/dix/dix_trace"
)

func init() {
	dix_trace.With(func(ctx *dix_trace.Ctx) {
		ctx.Func(Name, func() interface{} {
			var dt []string
			data.Each(func(key string) { dt = append(dt, key) })
			return dt
		})
	})
}
