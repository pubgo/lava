package xgen

import (
	"fmt"

	"github.com/pubgo/x/stack"

	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/vars"
)

func init() {
	vars.Watch("xgen", func() interface{} {
		var dt typex.Map
		for k, v := range List() {
			dt.Set(stack.Func(k.Interface()), fmt.Sprintf("%#v", v))
		}
		return dt.Map()
	})
}
