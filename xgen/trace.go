package xgen

import (
	"github.com/pubgo/lug/vars"
	"github.com/pubgo/x/stack"
)

func init() {
	vars.Watch("xgen", func() interface{} {
		var dt []interface{}
		for k := range List() {
			dt = append(dt, stack.Func(k.Interface()))
		}
		return dt
	})
}
