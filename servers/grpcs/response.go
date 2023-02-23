package grpcs

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lava/service"
	"google.golang.org/grpc"
)

var _ service.Response = (*rpcResponse)(nil)

type rpcResponse struct {
	stream grpc.ServerStream
	header *service.ResponseHeader
	dt     interface{}
}

func (h *rpcResponse) Header() *service.ResponseHeader { return h.header }
func (h *rpcResponse) Payload() interface{}            { return h.dt }
func (h *rpcResponse) Stream() bool                    { return h.stream != nil }

var _ service.Response = (*httpResponse)(nil)

type httpResponse struct {
	ctx *fiber.Ctx
}

func (h *httpResponse) Header() *service.ResponseHeader { return &h.ctx.Response().Header }
func (h *httpResponse) Payload() interface{}            { return h.ctx.Response().Body() }
func (h *httpResponse) Stream() bool                    { return false }
