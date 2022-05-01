package orm

import (
	"github.com/pubgo/xerror"
	"gorm.io/gorm"

	"github.com/pubgo/lava/config"
)

type Factory func(cfg config.CfgMap) gorm.Dialector

var factories = make(map[string]Factory)

func Get(name string) Factory  { return factories[name] }
func List() map[string]Factory { return factories }
func Register(name string, broker Factory) {
	xerror.Assert(name == "" || broker == nil, "[broker,name] should not be null")
	xerror.Assert(factories[name] != nil, "[broker] %s already exists", name)
	factories[name] = broker
}
