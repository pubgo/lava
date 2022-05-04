package import_registry

import (
	"go.uber.org/fx"

	"github.com/pubgo/lava/core/registry/registry_builder"
	"github.com/pubgo/lava/inject"
)

func init() {
	inject.Register(fx.Invoke(registry_builder.Enable))
}
