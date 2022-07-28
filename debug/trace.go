package debug

import (
	"github.com/gofiber/adaptor/v2"
	"github.com/pubgo/funk/recovery"
	"golang.org/x/net/trace"
)

func init() {
	defer recovery.Exit()

	Get("/requests", adaptor.HTTPHandlerFunc(trace.Traces))
	Get("/events", adaptor.HTTPHandlerFunc(trace.Events))
}
