package xgen

import (
	"github.com/pubgo/golug/tracelog"
	"github.com/pubgo/xprocess/xutil"
)

func init() {
	tracelog.Watch("golug_xgen", func() interface{} {
		dt := make(map[string][]GrpcRestHandler)
		for k, v := range List() {
			dt[xutil.FuncStack(k.Interface())] = v
		}
		return dt
	})
}
