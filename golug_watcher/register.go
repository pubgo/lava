package golug_watcher

import (
	"github.com/pubgo/golug/golug_types"
	"github.com/pubgo/xerror"
)

var watcherMap = golug_types.NewSyncMap()

func Register(name string, w Watcher) {
	xerror.Assert(name == "" || w == nil, "[watcher:%s] should not be null", name)
	xerror.Assert(watcherMap.Has(name), "[watcher:%s] already exists", name)

	watcherMap.Set(name, w)
}

func List() (dt map[string]Watcher) {
	watcherMap.Map(&dt)
	return dt
}
