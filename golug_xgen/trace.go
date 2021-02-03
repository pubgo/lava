package golug_xgen

import (
	"github.com/pubgo/golug/golug_trace"
	"github.com/pubgo/xprocess/xutil"
)

func init() {
	golug_trace.Watch("xgen", func() interface{} {
		dt := make(map[string][]GrpcRestHandler)
		for k, v := range List() {
			dt[xutil.FuncStack(k.Interface())] = v
		}
		return dt
	})
}
