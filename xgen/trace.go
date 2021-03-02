package xgen

import (
	"github.com/pubgo/golug/tracelog"
	"github.com/pubgo/x/stack"
)

func init() {
	tracelog.Watch("golug_xgen", func() interface{} {
		dt := make(map[string][]GrpcRestHandler)
		for k, v := range List() {
			dt[stack.Func(k.Interface())] = v
		}
		return dt
	})
}
