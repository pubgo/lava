package grpcEntry

import (
	"encoding/json"

	"google.golang.org/grpc"

	"github.com/pubgo/lava/types"
)

var _ types.Response = (*rpcResponse)(nil)

type rpcResponse struct {
	stream grpc.ServerStream
	header types.Header
	dt     interface{}
}

func (h *rpcResponse) Header() types.Header {
	return h.header
}

func (h *rpcResponse) Body() ([]byte, error) {
	return json.Marshal(h.dt)
}

func (h *rpcResponse) Payload() interface{} {
	return h.dt
}

func (h *rpcResponse) Stream() bool {
	return h.stream != nil
}
