package service_builder

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lava/abc"
	"google.golang.org/grpc"
)

var _ abc.Response = (*rpcResponse)(nil)

type rpcResponse struct {
	stream grpc.ServerStream
	header *abc.ResponseHeader
	dt     interface{}
}

func (h *rpcResponse) Header() *abc.ResponseHeader { return h.header }
func (h *rpcResponse) Payload() interface{}        { return h.dt }
func (h *rpcResponse) Stream() bool                { return h.stream != nil }

var _ abc.Response = (*httpResponse)(nil)

type httpResponse struct {
	ctx *fiber.Ctx
}

func (h *httpResponse) Write(p []byte) (n int, err error) { return h.ctx.Write(p) }
func (h *httpResponse) Header() *abc.ResponseHeader       { return &h.ctx.Response().Header }
func (h *httpResponse) Payload() interface{}              { return h.ctx.Response().Body() }
func (h *httpResponse) Stream() bool                      { return false }
