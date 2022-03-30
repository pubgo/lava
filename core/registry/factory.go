package registry

import (
	"github.com/pubgo/lava/config/config_type"
	"github.com/pubgo/lava/core/registry/registry_type"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/xerror"
)

type Builder func(cfgMap config_type.CfgMap) (registry_type.Registry, error)

var builders typex.Map

func Register(name string, r Builder) {
	defer xerror.RespExit()
	xerror.Assert(name == "" || r == nil, "[name,r] should not be null")
	xerror.Assert(builders.Has(name), "registry %s already exists", name)
	builders.Set(name, r)
}
