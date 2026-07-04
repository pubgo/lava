package gatewayserver

import (
	"google.golang.org/grpc"

	"github.com/pubgo/lava/v2/lava"
)

var _ lava.Response = (*rpcResponse)(nil)

type rpcResponse struct {
	stream grpc.ServerStream
	header *lava.ResponseHeader
	dt     any
}

func (h *rpcResponse) Header() *lava.ResponseHeader { return h.header }
func (h *rpcResponse) Payload() any                 { return h.dt }
func (h *rpcResponse) Stream() bool                 { return h.stream != nil }
