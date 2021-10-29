package ginEntry

import (
	"github.com/pubgo/lava/vars"
	"github.com/pubgo/x/stack"
)

func trace(t *ginEntry) {
	vars.Watch(t.Options().Name+"_cfg", func() interface{} { return t.cfg })
	vars.Watch(t.Options().Name+"_gin_router", func() interface{} {
		if t.srv == nil {
			return nil
		}

		var data = make(map[string]string)
		for _, r := range t.srv.Routes() {
			data[r.Method+" "+r.Path] = func() string {
				if r.Handler != "" {
					return r.Handler
				}

				return stack.Func(r.HandlerFunc)
			}()
		}
		return data
	})
}
