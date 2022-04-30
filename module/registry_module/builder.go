package registry_module

import (
	"github.com/pubgo/lava/inject"
	"go.uber.org/fx"

	"github.com/pubgo/lava/registry/registry_builder"
)

func init() {
	inject.Register(fx.Invoke(registry_builder.Enable))
}
