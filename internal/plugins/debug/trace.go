package debug

import (
	"golang.org/x/net/trace"

	"github.com/pubgo/lava/mux"
)

func init() {
	mux.DebugGet("/requests", trace.Traces)
	mux.DebugGet("/events", trace.Events)
}
