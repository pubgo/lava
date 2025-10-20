package healthy

import (
	"github.com/pubgo/funk/v2/stack"
	"github.com/pubgo/funk/v2/vars"
)

func init() {
	vars.Register(Name, func() any {
		data := make(map[string]any)
		healthList.Range(func(key, value any) bool {
			data[key.(string)] = stack.CallerWithFunc(value)
			return true
		})
		return data
	})
}
