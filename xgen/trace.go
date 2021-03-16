package xgen

import (
	"github.com/pubgo/golug/tracelog"
	"github.com/pubgo/x/stack"
)

func init() {
	tracelog.Watch("xgen", func() interface{} {
		var dt []interface{}
		for k := range List() {
			dt = append(dt, stack.Func(k.Interface()))
		}
		return dt
	})
}
