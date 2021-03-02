package watcher

import (
	"github.com/pubgo/golug/tracelog"
	"github.com/pubgo/xprocess/xutil"
)

func init() {
	tracelog.Watch(Name+"_watcher_callback", func() interface{} {
		var dt []string
		callbacks.Each(func(key string) { dt = append(dt, key) })
		return dt
	})

	tracelog.Watch(Name+"_watcher", func() interface{} {
		var dt = make(map[string]string)
		for k, v := range List() {
			dt[k] = xutil.FuncStack(v)
		}
		return dt
	})
}
