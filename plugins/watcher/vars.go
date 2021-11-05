package watcher

import (
	"github.com/pubgo/lava/vars"
	"github.com/pubgo/x/stack"
)

func init() {
	vars.Watch(Name+"_factories", func() interface{} {
		var data = make(map[string]string)
		for k, v := range factories {
			data[k] = stack.Func(v)
		}
		return data
	})

	vars.Watch(Name+"_callbacks", func() interface{} {
		var data = make(map[string][]string)
		for k, v := range callbacks {
			for i := range v {
				data[k] = append(data[k], stack.Func(v[i]))
			}
		}
		return data
	})
}
