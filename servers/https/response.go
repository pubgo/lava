package https

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lava/service"
)

var _ service.Response = (*httpResponse)(nil)

type httpResponse struct {
	ctx *fiber.Ctx
}

func (h *httpResponse) Header() *service.ResponseHeader { return &h.ctx.Response().Header }
func (h *httpResponse) Payload() interface{}            { return h.ctx.Response().Body() }
func (h *httpResponse) Stream() bool                    { return false }
