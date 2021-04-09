package registry

import (
	"github.com/pubgo/lug/types"
	"github.com/pubgo/xerror"
)

var factories types.SMap
var Default Registry

func Register(name string, r Factory) {
	defer xerror.RespExit()
	xerror.Assert(name == "" || r == nil, "[name,r] should not be null")
	xerror.Assert(factories.Has(name), "registry %s already exists", name)
	factories.Set(name, r)
}
