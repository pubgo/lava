package panicparse

import (
	"github.com/maruel/panicparse/v2/stack/webstack"

	"github.com/pubgo/lug/internal/debug"
	"github.com/pubgo/lug/types"
)

func init() {
	debug.On(func(mux *types.DebugMux) {
		mux.Get("/debug/panicparse",webstack.SnapshotHandler)
	})
}
