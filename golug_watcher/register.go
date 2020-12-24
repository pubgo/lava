package golug_watcher

import (
	"sync"

	"github.com/pubgo/xerror"
)

var registerMap sync.Map

func Register(name string, w func() Watcher) {
	if w == nil {
		xerror.Next().Panic(xerror.Fmt("[watcher] %s is nil", name))
	}

	if _, ok := registerMap.LoadOrStore(name, w); ok {
		xerror.Next().Panic(xerror.Fmt("[watcher] %s already exists", name))
	}
}

func List() map[string]func() Watcher {
	var dt = make(map[string]func() Watcher)
	registerMap.Range(func(key, value interface{}) bool {
		dt[key.(string)] = value.(func() Watcher)
		return true
	})
	return dt
}
