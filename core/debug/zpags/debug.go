package zpags

import (
	"github.com/gofiber/adaptor/v2"
	"github.com/pubgo/lava/core/debug"
	"go.opentelemetry.io/contrib/zpages"
)

func init() {
	debug.Get("/z", adaptor.HTTPHandler(zpages.NewTracezHandler(zpages.NewSpanProcessor())))
}
