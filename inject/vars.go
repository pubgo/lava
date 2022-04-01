package inject

import (
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/vars"
	"github.com/pubgo/x/stack"
)

func init() {
	vars.Register("inject", func() interface{} {
		var data typex.A
		for k, v := range injectHandlers {
			data.Append(typex.Kv{Key: k.String(), Value: stack.Func(v)})
		}
		return data
	})
}
