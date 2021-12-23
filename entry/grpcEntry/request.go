package grpcEntry

import (
	"google.golang.org/grpc"

	"github.com/pubgo/lava/encoding"
	"github.com/pubgo/lava/types"
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
	cdc           encoding.Codec
	header        types.Header
	payload       interface{}
}

func (r *rpcRequest) Operation() string {
	return r.method
}

func (r *rpcRequest) Kind() string         { return Name }
func (r *rpcRequest) Client() bool         { return false }
func (r *rpcRequest) Header() types.Header { return r.header }
func (r *rpcRequest) Payload() interface{} { return r.payload }
func (r *rpcRequest) ContentType() string  { return r.contentType }
func (r *rpcRequest) Service() string      { return r.service }
func (r *rpcRequest) Endpoint() string     { return r.method }
func (r *rpcRequest) Stream() bool         { return r.stream != nil }
