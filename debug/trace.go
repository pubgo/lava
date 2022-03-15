package debug

import (
	"github.com/pubgo/lava/debug/debug_mux"
	"golang.org/x/net/trace"
)

func init() {
	debug_mux.DebugGet("/requests", trace.Traces)
	debug_mux.DebugGet("/events", trace.Events)
}
