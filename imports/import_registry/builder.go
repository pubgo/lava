package import_registry

import (
	"github.com/pubgo/lava/inject"
	"go.uber.org/fx"

	"github.com/pubgo/lava/core/registry/registry_builder"
)

func init() {
	inject.Register(fx.Invoke(registry_builder.Enable))
}
