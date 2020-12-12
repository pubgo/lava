package golug_watcher

import (
	"sync"

	"github.com/pubgo/xerror"
)

var register sync.Map

func Register(name string, w Watcher) {
	if w == nil {
		xerror.Next().Panic(xerror.Fmt("[watcher] %s is nil", name))
	}
	register.Store(name, w)
}

func List() map[string]Watcher {
	var dt = make(map[string]Watcher)
	register.Range(func(key, value interface{}) bool {
		dt[key.(string)] = value.(Watcher)
		return true
	})
	return dt
}
