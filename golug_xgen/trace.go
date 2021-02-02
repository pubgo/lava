package golug_xgen

import (
	"github.com/pubgo/dix/dix_trace"
	"github.com/pubgo/xprocess/xutil"
)

func init() {
	dix_trace.With(func(ctx *dix_trace.Ctx) {
		ctx.Func("xgen", func() interface{} {
			dt := make(map[string][]GrpcRestHandler)
			for k, v := range List() {
				dt[xutil.FuncStack(k.Interface())] = v
			}
			return dt
		})
	})
}
