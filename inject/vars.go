package inject

import (
	"reflect"

	"github.com/pubgo/lava/vars"
	"github.com/pubgo/x/stack"
)

func init() {
	vars.Register("inject", func() interface{} {
		var data = make(map[reflect.Type]string)
		for k, v := range injectHandlers {
			data[k] = stack.Func(v)
		}
		return data
	})
}
