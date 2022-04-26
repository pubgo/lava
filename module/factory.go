package module

import (
	"github.com/pubgo/xerror"
	"go.uber.org/fx"
)

var factories = make(map[string]fx.Option)

func Get(name string) fx.Option  { return factories[name] }
func List() map[string]fx.Option { return factories }
func Register(name string, m fx.Option) {
	xerror.Assert(name == "" || m == nil, "[m,name] should not be null")
	xerror.Assert(factories[name] != nil, "[name] %s already exists")
	factories[name] = m
}
