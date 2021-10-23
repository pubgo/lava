package ginEntry

import (
	"github.com/pubgo/lava/vars"
)

func trace(t *restEntry) {
	vars.Watch(t.Options().Name+"_cfg", func() interface{} { return t.cfg })
	vars.Watch(t.Options().Name+"_rest_router", func() interface{} {
		if t.srv.Get() == nil {
			return nil
		}

		var data = make(map[string][]string)
		stack := t.srv.Get().Stack()
		for m := range stack {
			for _, route := range stack[m] {
				data[route.Path] = append(data[route.Path], route.Method)
			}
		}
		return data
	})
}
