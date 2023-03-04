package healthy

import (
	"github.com/pubgo/funk/stack"
	"github.com/pubgo/lava/core/vars"
)

func init() {
	vars.Register(Name, func() interface{} {
		var data = make(map[string]any)
		healthList.Range(func(key, value interface{}) bool {
			data[key.(string)] = stack.CallerWithFunc(value)
			return true
		})
		return data
	})
}
