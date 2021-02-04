package golug_watcher

import (
	"github.com/pubgo/golug/golug_types"
	"github.com/pubgo/xerror"
)

var watchers golug_types.SyncMap

func List() (dt map[string]Watcher) { watchers.Map(&dt); return dt }
func Register(name string, w Watcher) {
	xerror.Assert(name == "" || w == nil, "[watcher:%s] should not be null", name)
	xerror.Assert(watchers.Has(name), "[watcher:%s] already exists", name)

	watchers.Set(name, w)
}
