package debug_module

import (
	"github.com/pubgo/lava/inject"
	"go.uber.org/fx"

	"github.com/pubgo/lava/debug/debug_srv"
)

func init() {
	inject.Register(fx.Invoke(debug_srv.Enable))
}
