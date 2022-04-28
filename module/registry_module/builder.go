package registry_module

import (
	"go.uber.org/fx"

	"github.com/pubgo/lava/module"
	"github.com/pubgo/lava/registry/registry_builder"
)

func init() {
	module.Register(fx.Invoke(registry_builder.Enable))
}
