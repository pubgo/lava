package xgen

import (
	"github.com/pubgo/golug/tracelog"
	"github.com/pubgo/x/stack"
)

func init() {
	tracelog.Watch("xgen", func() interface{} {
		dt := make(map[string]struct{})
		for k := range List() {
			dt[stack.Func(k.Interface())] = struct{}{}
		}
		return dt
	})
}
