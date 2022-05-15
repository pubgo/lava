package import_registry

import (
	"github.com/pubgo/lava/core/registry/registry_builder"
	"github.com/pubgo/lava/inject"
)

func init() {
	inject.Invoke(registry_builder.Enable)
}
