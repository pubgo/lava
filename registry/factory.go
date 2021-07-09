package registry

import (
	"github.com/pubgo/lug/pkg/typex"
	"github.com/pubgo/xerror"
)

type Factory func(map[string]interface{}) (Registry, error)

var factories typex.Map

func Register(name string, r Factory) {
	defer xerror.RespExit()
	
	xerror.Assert(name == "" || r == nil, "[name,r] should not be null")
	xerror.Assert(factories.Has(name), "registry %s already exists", name)
	factories.Set(name, r)
}
