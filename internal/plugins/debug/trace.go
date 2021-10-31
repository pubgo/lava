package debug

import (
	"github.com/pubgo/lava/mux"
	"golang.org/x/net/trace"
)

func init() {
	mux.Get("/debug/requests", trace.Traces)
	mux.Get("/debug/events", trace.Events)
}
