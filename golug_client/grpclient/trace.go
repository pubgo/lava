package grpclient

import (
	"github.com/pubgo/dix/dix_trace"
)

func init() {
	dix_trace.With(func(ctx *dix_trace.TraceCtx) {
		ctx.Func(Name+"_cfg", func() interface{} { return cfg })
	})
}
