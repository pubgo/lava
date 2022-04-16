package tracing

import (
	"github.com/pubgo/lava/vars"
	"github.com/pubgo/x/stack"
)

func init() {
	vars.Register(Name+"_factory", func() interface{} {
		var data = make(map[string]string)
		factories.Range(func(key, value interface{}) bool {
			data[key.(string)] = stack.Func(value)
			return true
		})
		return data
	})
}