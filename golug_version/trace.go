package golug_version

import (
	"github.com/pubgo/dix/dix_trace"
)

func init() {
	dix_trace.With(func(ctx *dix_trace.Ctx) {
		ctx.Func("golug_version", func() interface{} { return List() })
	})
}
