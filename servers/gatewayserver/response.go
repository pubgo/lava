package gatewayserver

import (
	"github.com/pubgo/lava/v2/pkg/lava"
)

var _ lava.Response = (*rpcResponse)(nil)

type rpcResponse struct {
	header   lava.ResponseHeader
	dt       any
	streamed bool
}

func (h *rpcResponse) Header() lava.ResponseHeader { return h.header }
func (h *rpcResponse) Payload() any                { return h.dt }
func (h *rpcResponse) Stream() bool                { return h.streamed }
