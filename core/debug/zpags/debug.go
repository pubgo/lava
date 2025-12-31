package zpags

import (
	"github.com/gofiber/adaptor/v2"
	"github.com/pubgo/lava/v2/core/debug"
	"go.opentelemetry.io/contrib/zpages"
)

func init() {
	debug.Get("/z", adaptor.HTTPHandler(zpages.NewTracezHandler(zpages.NewSpanProcessor())))
}
