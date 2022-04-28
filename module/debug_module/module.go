package debug_module

import (
	"github.com/pubgo/lava/debug/debug_srv"
	"github.com/pubgo/lava/module"
)

func init() {
	module.Invoke(debug_srv.Enable)
}
