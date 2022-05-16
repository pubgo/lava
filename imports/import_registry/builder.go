package import_registry

import (
	"github.com/pubgo/lava/core/registry"
	"github.com/pubgo/lava/inject"
)

func init() {
	inject.Invoke(registry.Enable)
}
