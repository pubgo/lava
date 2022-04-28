package debug_module

import (
	"go.uber.org/fx"

	"github.com/pubgo/lava/debug/debug_srv"
	"github.com/pubgo/lava/module"
)

func init() {
	module.Register(fx.Invoke(debug_srv.Enable))
}
