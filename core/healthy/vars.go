package healthy

import (
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/lava/vars"
	"github.com/pubgo/x/stack"
)

func init() {
	defer recovery.Exit()

	vars.Register(Name, func() interface{} {
		var data = make(map[string]string)
		healthList.Range(func(key, value interface{}) bool {
			data[key.(string)] = stack.Func(value)
			return true
		})
		return data
	})
}
