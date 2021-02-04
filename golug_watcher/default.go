package golug_watcher

import (
	"github.com/pubgo/golug/golug_types"
	"github.com/pubgo/xerror"
)

var callbacks = golug_types.NewSyncMap()

func Watch(name string, h CallBack) {
	xerror.Assert(name == "" || h == nil, "[name], [callback] should not be null")
	xerror.Assert(callbacks.Has(name), "[callback] %s already exists", name)

	callbacks.Set(name, h)
}

func GetWatch(name string) CallBack {
	val, ok := callbacks.Load(name)
	if !ok {
		return nil
	}
	return val.(CallBack)
}
