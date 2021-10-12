package grpcEntry

import (
	"encoding/json"

	"github.com/pubgo/lava/types"

	"google.golang.org/grpc"
)

var _ types.Response = (*rpcResponse)(nil)

type rpcResponse struct {
	stream grpc.ServerStream
	header types.Header
	dt     interface{}
	ct     string
}

func (h *rpcResponse) Write(p []byte) (n int, err error) {
	panic("implement me")
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

func (h *rpcResponse) Codec() string {
	return h.ct
}

func (h *rpcResponse) Send(i interface{}) error {
	return h.stream.SendMsg(i)
}

func (h *rpcResponse) Recv(i interface{}) error {
	return h.stream.RecvMsg(i)
}
