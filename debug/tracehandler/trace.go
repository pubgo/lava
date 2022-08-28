package tracehandler

import (
	"github.com/gofiber/adaptor/v2"
	"github.com/pubgo/lava/debug"
	"golang.org/x/net/trace"
)

func init() {
	debug.Get("/requests", adaptor.HTTPHandlerFunc(trace.Traces))
	debug.Get("/events", adaptor.HTTPHandlerFunc(trace.Events))
}
