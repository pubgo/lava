package golug_watcher

import (
	"sync"

	"github.com/pubgo/xerror"
)

var callbackMap sync.Map

func Watch(name string, h CallBack) {
	xerror.Assert(name == "" || h == nil, "[name], [callback] should not be null")

	if _, ok := callbackMap.LoadOrStore(name, h); ok {
		xerror.Assert(ok, "[callback] %s already exists", name)
	}
}

func GetWatch(name string) CallBack {
	val, ok := callbackMap.Load(name)
	if !ok {
		return nil
	}
	return val.(CallBack)
}
