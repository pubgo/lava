package golug_log

import (
	"github.com/pubgo/dix/dix_trace"
)

func init() {
	dix_trace.With(func(ctx *dix_trace.Ctx) { ctx.Func(Name, func() interface{} { return cfg }) })
}
