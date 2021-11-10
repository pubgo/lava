package orm

import (
	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
	"gorm.io/gorm"

	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/types"
)

type Factory func(cfg types.CfgMap) (gorm.Dialector, error)

var factories typex.SMap

func List() (dt map[string]Factory) { xerror.Panic(factories.MapTo(&dt)); return }
func Register(name string, broker Factory) {
	xerror.Assert(name == "" || broker == nil, "[broker,name] should not be null")
	xerror.Assert(factories.Has(name), "[broker] %s already exists, refer: %s", name, stack.Func(factories.Get(name)))
	factories.Set(name, broker)
}
