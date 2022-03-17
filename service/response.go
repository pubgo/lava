package service

import (
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/types"
)

var _ types.Response = (*rpcResponse)(nil)

type rpcResponse struct {
	stream grpc.ServerStream
	header types.Header
	dt     interface{}
}

func (h *rpcResponse) Header() types.Header { return h.header }
func (h *rpcResponse) Payload() interface{} { return h.dt }
func (h *rpcResponse) Stream() bool         { return h.stream != nil }

var _ types.Response = (*httpResponse)(nil)

type httpResponse struct {
	ctx    *fiber.Ctx
	header types.Header
}

func (h *httpResponse) Write(p []byte) (n int, err error) {
	return h.ctx.Write(p)
}

func (h *httpResponse) Header() types.Header {
	return h.header
}

func (h *httpResponse) Payload() interface{} {
	return h.ctx.Response().Body()
}

func (h *httpResponse) Stream() bool {
	return false
}
