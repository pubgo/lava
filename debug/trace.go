package debug

import (
	"github.com/gofiber/adaptor/v2"
	"golang.org/x/net/trace"
)

func init() {
	Get("/requests", adaptor.HTTPHandlerFunc(trace.Traces))
	Get("/events", adaptor.HTTPHandlerFunc(trace.Events))
}
