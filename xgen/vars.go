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
			var mthList, ok = v.([]GrpcRestHandler)
			if !ok {
				dt.Set(stack.Func(k.Interface()), fmt.Sprintf("%#v", v))
				continue
			}

			var data1 []string
			for i := range mthList {
				data1 = append(data1, fmt.Sprintf("%#v", mthList[i]))
			}
			dt.Set(stack.Func(k.Interface()), data1)
		}
		return dt.Map()
	})
}
