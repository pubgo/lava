package healthy

import (
	"github.com/pubgo/funk/stack"
	"github.com/pubgo/funk/vars"
)

func init() {
	vars.Register(Name, func() interface{} {
		data := make(map[string]any)
		healthList.Range(func(key, value interface{}) bool {
			data[key.(string)] = stack.CallerWithFunc(value)
			return true
		})
		return data
	})
}
