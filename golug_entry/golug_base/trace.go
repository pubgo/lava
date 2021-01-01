package golug_base

import (
	"github.com/pubgo/dix/dix_trace"
)

func (t *baseEntry) trace() {
	dix_trace.With(func(ctx *dix_trace.Ctx) {
		ctx.Func(t.Options().Name, func() interface{} { return t.cfg })
	})
}
