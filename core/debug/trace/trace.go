package trace

import (
	adaptor "github.com/gofiber/fiber/v3/middleware/adaptor"
	"golang.org/x/net/trace"

	"github.com/pubgo/lava/core/debug"
)

func init() {
	debug.Get("/requests", adaptor.HTTPHandlerFunc(trace.Traces))
	debug.Get("/events", adaptor.HTTPHandlerFunc(trace.Events))
}
