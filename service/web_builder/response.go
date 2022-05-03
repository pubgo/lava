package service_builder

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lava/middleware"
)

var _ middleware.Response = (*httpResponse)(nil)

type httpResponse struct {
	ctx *fiber.Ctx
}

func (h *httpResponse) Write(p []byte) (n int, err error)  { return h.ctx.Write(p) }
func (h *httpResponse) Header() *middleware.ResponseHeader { return &h.ctx.Response().Header }
func (h *httpResponse) Payload() interface{}               { return h.ctx.Response().Body() }
func (h *httpResponse) Stream() bool                       { return false }
