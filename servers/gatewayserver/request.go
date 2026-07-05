package gatewayserver

import (
	"google.golang.org/grpc"

	"github.com/pubgo/lava/v2/pkg/lava"
)

var _ lava.Request = (*rpcRequest)(nil)

type rpcRequest struct {
	handler       grpc.UnaryHandler
	handlerStream grpc.StreamHandler
	stream        grpc.ServerStream
	srv           any
	service       string
	method        string
	url           string
	contentType   string
	header        lava.RequestHeader
	rspHeader     lava.ResponseHeader
	payload       any
}

func (r *rpcRequest) Kind() string               { return lava.RequestKindGrpc }
func (r *rpcRequest) Client() bool               { return false }
func (r *rpcRequest) Header() lava.RequestHeader { return r.header }
func (r *rpcRequest) Payload() any               { return r.payload }
func (r *rpcRequest) ContentType() string        { return r.contentType }
func (r *rpcRequest) Service() string            { return r.service }
func (r *rpcRequest) Operation() string          { return r.method }
func (r *rpcRequest) Endpoint() string           { return r.url }
func (r *rpcRequest) Stream() bool               { return r.stream != nil }
