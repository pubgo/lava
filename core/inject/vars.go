package inject

import (
	"expvar"

	"github.com/pubgo/x/stack"
)

func init() {
	expvar.Publish("inject", expvar.Func(func() interface{} {
		var data = make(map[string]string)
		for k, v := range injectHandlers {
			data[k.String()] = stack.Func(v)
		}
		return data
	}))
}
