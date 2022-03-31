package service_builder

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lava/service"
	"google.golang.org/grpc"
)

var _ service.Response = (*rpcResponse)(nil)

type rpcResponse struct {
	stream grpc.ServerStream
	header service.Header
	dt     interface{}
}

func (h *rpcResponse) Header() service.Header { return h.header }
func (h *rpcResponse) Payload() interface{}   { return h.dt }
func (h *rpcResponse) Stream() bool           { return h.stream != nil }

var _ service.Response = (*httpResponse)(nil)

type httpResponse struct {
	ctx    *fiber.Ctx
	header service.Header
}

func (h *httpResponse) Write(p []byte) (n int, err error) {
	return h.ctx.Write(p)
}

func (h *httpResponse) Header() service.Header {
	return h.header
}

func (h *httpResponse) Payload() interface{} {
	return h.ctx.Response().Body()
}

func (h *httpResponse) Stream() bool {
	return false
}
