package watcher

import (
	"github.com/pubgo/lug/vars"
	"github.com/pubgo/x/stack"
)

func init() {
	vars.Watch(Name+"_callback", func() interface{} {
		var dt []string
		callbacks.Each(func(key string, _ interface{}) { dt = append(dt, key) })
		return dt
	})

	vars.Watch(Name, func() interface{} {
		var dt = make(map[string]string)
		for name, f := range factories {
			dt[name] = stack.Func(f)
		}
		return dt
	})
}
