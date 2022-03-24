package watcher

import (
	"github.com/pubgo/lava/core/watcher/watcher_type"
	"github.com/pubgo/lava/vars"
	"github.com/pubgo/x/stack"
)

func init() {
	vars.Register(watcher_type.Name+"_factories", func() interface{} {
		var data = make(map[string]string)
		for k, v := range factories {
			data[k] = stack.Func(v)
		}
		return data
	})

	vars.Register(watcher_type.Name+"_handlers", func() interface{} {
		var data = make(map[string][]string)
		for k, v := range callbacks {
			for i := range v {
				data[k] = append(data[k], stack.Func(v[i]))
			}
		}
		return data
	})
}
