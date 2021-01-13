package golug_watcher

import (
	"sync"

	"github.com/pubgo/xerror"
)

var watcherMap sync.Map

func Register(name string, w Watcher) {
	xerror.Assert(name == "" || w == nil, "[name], [watcher] should not be null", name)

	if _, ok := watcherMap.LoadOrStore(name, w); ok {
		xerror.Assert(ok, "[watcher] %s already exists", name)
	}
}

func List() map[string]Watcher {
	var dt = make(map[string]Watcher)
	watcherMap.Range(func(key, value interface{}) bool {
		dt[key.(string)] = value.(Watcher)
		return true
	})
	return dt
}
