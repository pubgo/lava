package golug_watcher

import (
	"sync"

	"github.com/pubgo/xerror"
)

var registerMap sync.Map

func Register(name string, w Watcher) {
	if w == nil {
		xerror.Next().Panic(xerror.Fmt("[watcher] %s is nil", name))
	}

	if _, ok := registerMap.LoadOrStore(name, w); ok {
		xerror.Next().Panic(xerror.Fmt("[watcher] %s already exists", name))
	}
}

func List() map[string]Watcher {
	var dt = make(map[string]Watcher)
	registerMap.Range(func(key, value interface{}) bool {
		if value.(func() Watcher) == nil {
			return true
		}

		dt[key.(string)] = value.(Watcher)
		return true
	})
	return dt
}
