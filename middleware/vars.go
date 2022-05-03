package middleware

import (
	"github.com/pubgo/lava/vars"
	"github.com/pubgo/x/stack"
)

func init() {
	vars.Register("middleware", func() interface{} {
		var data = make(map[string]string)
		for k, v := range factories {
			data[k] = stack.Func(v)
		}
		return data
	})
}
