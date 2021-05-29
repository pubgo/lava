package xgen

import (
	"github.com/pubgo/lug/pkg/typex"
	"github.com/pubgo/lug/vars"
	"github.com/pubgo/x/stack"
)

func init() {
	vars.Watch("xgen", func() interface{} {
		var dt typex.Map
		for k, v := range List() {
			dt.Set(stack.Func(k.Interface()), v)
		}
		return dt.Map()
	})
}
