package service

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lava/service/service_type"
	"google.golang.org/grpc"
)

var _ service_type.Response = (*rpcResponse)(nil)

type rpcResponse struct {
	stream grpc.ServerStream
	header Header
	dt     interface{}
}

func (h *rpcResponse) Header() Header       { return h.header }
func (h *rpcResponse) Payload() interface{} { return h.dt }
func (h *rpcResponse) Stream() bool         { return h.stream != nil }

var _ service_type.Response = (*httpResponse)(nil)

type httpResponse struct {
	ctx    *fiber.Ctx
	header Header
}

func (h *httpResponse) Write(p []byte) (n int, err error) {
	return h.ctx.Write(p)
}

func (h *httpResponse) Header() Header {
	return h.header
}

func (h *httpResponse) Payload() interface{} {
	return h.ctx.Response().Body()
}

func (h *httpResponse) Stream() bool {
	return false
}
