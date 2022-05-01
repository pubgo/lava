package debug_srv

import (
	"github.com/gofiber/adaptor/v2"
	"golang.org/x/net/trace"

	"github.com/pubgo/lava/debug"
)

func init() {
	debug.Get("/requests", adaptor.HTTPHandlerFunc(trace.Traces))
	debug.Get("/events", adaptor.HTTPHandlerFunc(trace.Events))
}
