package watcher

import (
	"github.com/pubgo/lug/types"
	"github.com/pubgo/xerror"
)

var factories types.SMap

func Register(name string, w Factory) {
	xerror.Assert(name == "" || w == nil, "[name,w] should not be null")
	xerror.Assert(factories.Has(name), "[name] %s already exists", name)
	factories.Set(name, w)
}
