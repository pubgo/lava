package grpc

import (
	"github.com/pubgo/lug/encoding"
	"github.com/pubgo/lug/types"

	"google.golang.org/grpc"
)

var _ types.Request = (*rpcRequest)(nil)

type rpcRequest struct {
	handler       grpc.UnaryHandler
	handlerStream grpc.StreamHandler
	stream        grpc.ServerStream
	srv           interface{}
	service       string
	method        string
	contentType   string
	cdc           string
	header        types.Header
	payload       interface{}
}

func (r *rpcRequest) Client() bool {
	return false
}

func (r *rpcRequest) Header() types.Header {
	return r.header
}

func (r *rpcRequest) Payload() interface{} {
	return r.payload
}

func (r *rpcRequest) Body() ([]byte, error) {
	var cdc = encoding.Get(r.cdc)
	if cdc == nil {
		return nil, encoding.ErrNotFound
	}

	return cdc.Marshal(r.payload)
}

func (r *rpcRequest) ContentType() string {
	return r.contentType
}

func (r *rpcRequest) Service() string {
	return r.service
}

func (r *rpcRequest) Method() string {
	return r.method
}

func (r *rpcRequest) Endpoint() string {
	return r.method
}

func (r *rpcRequest) Codec() string {
	return r.cdc
}

func (r *rpcRequest) Stream() bool {
	return r.stream != nil
}
