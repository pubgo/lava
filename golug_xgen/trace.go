package golug_xgen

import (
	"reflect"
	
	"github.com/pubgo/dix/dix_trace"
	"github.com/pubgo/xerror/xerror_util"
)

func init() {
	dix_trace.With(func(ctx *dix_trace.TraceCtx) {
		ctx.Func("xgen", func() interface{} {
			dt := make(map[string][]GrpcRestHandler)
			data.Range(func(key, value interface{}) bool {
				dt[xerror_util.CallerWithFunc(key.(reflect.Value).Interface())] = value.([]GrpcRestHandler)
				return true
			})
			return dt
		})
	})
}
