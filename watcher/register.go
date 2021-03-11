package watcher

import (
	"github.com/pubgo/golug/types"
	"github.com/pubgo/xerror"
)

var factories types.SMap

func Register(name string, w Factory) {
	xerror.Assert(name == "" || w == nil, "[watcher:%s] should not be null", name)
	xerror.Assert(factories.Has(name), "[watcher:%s] already exists", name)
	factories.Set(name, w)
}
