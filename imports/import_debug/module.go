package import_debug

import (
	"go.uber.org/fx"

	"github.com/pubgo/lava/debug/debug_srv"
	"github.com/pubgo/lava/inject"
)

func init() {
	inject.Register(fx.Invoke(debug_srv.Enable))
}
