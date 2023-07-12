package grpcs

import (
	"google.golang.org/grpc"

	"github.com/pubgo/lava/lava"
)

var _ lava.Response = (*rpcResponse)(nil)

type rpcResponse struct {
	stream grpc.ServerStream
	header *lava.ResponseHeader
	dt     interface{}
}

func (h *rpcResponse) Header() *lava.ResponseHeader { return h.header }
func (h *rpcResponse) Payload() interface{}         { return h.dt }
func (h *rpcResponse) Stream() bool                 { return h.stream != nil }
