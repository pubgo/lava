package healthy

import (
	"github.com/pubgo/lava/vars"
	"github.com/pubgo/x/stack"
)

func init() {
	vars.Watch(Name, func() interface{} {
		var data = make(map[string]string)
		healthList.Range(func(key, value interface{}) bool {
			data[key.(string)] = stack.Func(value)
			return true
		})
		return data
	})
}
