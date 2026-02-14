package zpags

import (
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"go.opentelemetry.io/contrib/zpages"

	"github.com/pubgo/lava/v2/core/debug"
)

func init() {
	debug.Get("/z", adaptor.HTTPHandler(zpages.NewTracezHandler(zpages.NewSpanProcessor())))
}
