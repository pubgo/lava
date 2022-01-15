package metric

import (
	"github.com/pubgo/xerror"
	"github.com/uber-go/tally"

	"github.com/pubgo/lava/pkg/lavax"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/types"
)

type Factory func(cfg types.CfgMap, opts *tally.ScopeOptions) error

var factories typex.SMap

func GetFactory(names ...string) Factory {
	val, ok := factories.Load(lavax.GetDefault(names...))
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
