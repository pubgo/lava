package metric

import (
	"github.com/pubgo/lava/config/config_type"
	"github.com/pubgo/xerror"
	"github.com/uber-go/tally"

	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/pkg/utils"
)

type Factory func(cfg config_type.CfgMap, opts *tally.ScopeOptions) error

var factories typex.SMap

func GetFactory(names ...string) Factory {
	val, ok := factories.Load(utils.GetDefault(names...))
	if !ok {
		return nil
	}

	return val.(Factory)
}

func RegisterFactory(name string, r Factory) {
	defer xerror.RespExit()
	xerror.Assert(name == "" || r == nil, "[name,r] is null")
	xerror.Assert(factories.Has(name), "reporter [%s] already exists", name)
	factories.Set(name, r)
	return
}
