package orm

import (
	"github.com/pubgo/xerror"
	"gorm.io/gorm"

	"github.com/pubgo/lava/config/config_type"
	"github.com/pubgo/lava/pkg/typex"
)

type Factory func(cfg config_type.CfgMap) gorm.Dialector

var factories typex.SMap

func List() (dt map[string]Factory) { xerror.Panic(factories.MapTo(&dt)); return }
func Register(name string, broker Factory) {
	defer xerror.RespExit()
	xerror.Assert(name == "" || broker == nil, "[broker,name] should not be null")
	xerror.Assert(factories.Has(name), "[broker] %s already exists", name)
	factories.Set(name, broker)
}
