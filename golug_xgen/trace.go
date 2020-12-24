package golug_xgen

import (
	"github.com/pubgo/dix/dix_trace"
	"github.com/pubgo/xerror/xerror_util"
)

func init() {
	dix_trace.With(func(ctx *dix_trace.TraceCtx) {
		ctx.Func("xgen", func() interface{} {
			dt := make(map[string][]GrpcRestHandler)
			for k, v := range List() {
				dt[xerror_util.CallerWithFunc(k.Interface())] = v
			}
			return dt
		})
	})
}
