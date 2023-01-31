package tracing

import (
	"github.com/pubgo/funk/recovery"
)

func init() {
	defer recovery.Exit()

	vars.Register(Name+"_factory", func() interface{} {
		var data = make(map[string]string)
		factories.Range(func(key, value interface{}) bool {
			data[key.(string)] = stack.Func(value)
			return true
		})
		return data
	})
}
