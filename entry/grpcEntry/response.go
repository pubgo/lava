package grpcEntry

import (
	"github.com/pubgo/lava/service"
	"github.com/pubgo/lava/service/service_type"
	"google.golang.org/grpc"
)

var _ service_type.Response = (*rpcResponse)(nil)

type rpcResponse struct {
	stream grpc.ServerStream
	header service.Header
	dt     interface{}
}

func (h *rpcResponse) Header() service.Header { return h.header }
func (h *rpcResponse) Payload() interface{}   { return h.dt }
func (h *rpcResponse) Stream() bool           { return h.stream != nil }
