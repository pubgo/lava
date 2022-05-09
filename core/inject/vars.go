package inject

import (
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/vars"
	"github.com/pubgo/x/stack"
)

func init() {
	vars.Register("inject", func() interface{} {
		var data typex.D
		for k, v := range typeProviders {
			data.Append(typex.Kv{K: k.String(), V: stack.Func(v)})
		}
		return data
	})
}
