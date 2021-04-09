package watcher

import (
	"github.com/pubgo/lug/vars"
	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
)

func init() {
	vars.Watch(Name+"_watcher_callback", func() interface{} {
		var dt []string
		xerror.Panic(callbacks.Each(func(key string) { dt = append(dt, key) }))
		return dt
	})

	vars.Watch(Name+"_watcher", func() interface{} {
		var dt = make(map[string]string)
		xerror.Panic(factories.Each(func(name string, f Factory) {
			dt[name] = stack.Func(f)
		}))
		return dt
	})
}
