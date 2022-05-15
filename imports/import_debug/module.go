package import_debug

import (
	"github.com/pubgo/lava/debug/debug_srv"
	"github.com/pubgo/lava/inject"
)

func init() {
	inject.Invoke(debug_srv.Enable)
}
