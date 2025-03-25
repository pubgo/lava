package panicparse

import (
	"github.com/maruel/panicparse/v2/stack/webstack"
	"github.com/pubgo/lava/core/debug"
)

func init() {
	debug.Get("/panicparse", debug.WrapFunc(webstack.SnapshotHandler))
}
