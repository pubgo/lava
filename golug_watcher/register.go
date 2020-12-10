package golug_watcher

import (
	"sync"

	"github.com/pubgo/xerror"
)

var register sync.Map

func Register(name string, w Watcher) {
	data.Store(name, w)
}

func Get(name string) Watcher {
	val, ok := register.Load(name)
	if ok {
		return val.(Watcher)
	}

	xerror.Next().Panic(xerror.Fmt("%s not found", name))
	return nil
}

func List() map[string]Watcher {
	var dt = make(map[string]Watcher)
	register.Range(func(key, value interface{}) bool {
		dt[key.(string)] = value.(Watcher)
		return true
	})
	return dt
}
