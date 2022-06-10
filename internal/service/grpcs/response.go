package grpcs

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lava/core/middleware"
	"google.golang.org/grpc"
)

var _ middleware.Response = (*rpcResponse)(nil)

type rpcResponse struct {
	stream grpc.ServerStream
	header *middleware.ResponseHeader
	dt     interface{}
}

func (h *rpcResponse) Header() *middleware.ResponseHeader { return h.header }
func (h *rpcResponse) Payload() interface{}               { return h.dt }
func (h *rpcResponse) Stream() bool                       { return h.stream != nil }

var _ middleware.Response = (*httpResponse)(nil)

type httpResponse struct {
	ctx *fiber.Ctx
}

func (h *httpResponse) Write(p []byte) (n int, err error)  { return h.ctx.Write(p) }
func (h *httpResponse) Header() *middleware.ResponseHeader { return &h.ctx.Response().Header }
func (h *httpResponse) Payload() interface{}               { return h.ctx.Response().Body() }
func (h *httpResponse) Stream() bool                       { return false }
