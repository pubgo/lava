package https

import (
	"github.com/gofiber/fiber/v3"

	"github.com/pubgo/lava/lava"
)

var _ lava.Response = (*httpResponse)(nil)

type httpResponse struct {
	ctx fiber.Ctx
}

func (h *httpResponse) Header() *lava.ResponseHeader { return &h.ctx.Response().Header }
func (h *httpResponse) Payload() interface{}         { return h.ctx.Response().Body() }
func (h *httpResponse) Stream() bool                 { return false }
