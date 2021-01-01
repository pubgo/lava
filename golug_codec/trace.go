package golug_codec

import (
	"github.com/pubgo/dix/dix_trace"
)

func init() {
	dix_trace.With(func(ctx *dix_trace.Ctx) {
		ctx.Func(Name, func() interface{} {
			var dt []string
			data.Range(func(key, value interface{}) bool {
				dt = append(dt, key.(string))
				return true
			})
			return dt
		})
	})
}
