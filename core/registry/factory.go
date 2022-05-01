package registry

import (
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/xerror"
)

type Builder func(cfgMap config.CfgMap) (Registry, error)

var builders typex.Map

func Register(name string, r Builder) {
	defer xerror.RespExit()
	xerror.Assert(name == "" || r == nil, "[name,r] should not be null")
	xerror.Assert(builders.Has(name), "registry %s already exists", name)
	builders.Set(name, r)
}
