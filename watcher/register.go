package watcher

import (
	"github.com/pubgo/golug/consts"
	"github.com/pubgo/golug/types"
	"github.com/pubgo/xerror"
)

var watchers types.SMap

func List() (dt map[string]Watcher) { xerror.Panic(watchers.Map(&dt)); return dt }
func Register(name string, w Watcher) {
	xerror.Assert(name == "" || w == nil || w.Name() == "", "[watcher:%s] should not be null", name)
	xerror.Assert(watchers.Has(name), "[watcher:%s] already exists", name)

	watchers.Set(name, w)
}

func Get(names ...string) Watcher {
	w := watchers.Get(consts.GetDefault(names...))
	if w == nil {
		return nil
	}

	return w.(Watcher)
}
