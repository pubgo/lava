package grpcs

import (
	"google.golang.org/grpc"

	"github.com/pubgo/lava/lava"
)

var _ lava.Request = (*rpcRequest)(nil)

type rpcRequest struct {
	handler       grpc.UnaryHandler
	handlerStream grpc.StreamHandler
	stream        grpc.ServerStream
	srv           interface{}
	service       string
	method        string
	url           string
	contentType   string
	header        *lava.RequestHeader
	payload       interface{}
}

func (r *rpcRequest) Kind() string                { return "grpc" }
func (r *rpcRequest) Client() bool                { return false }
func (r *rpcRequest) Header() *lava.RequestHeader { return r.header }
func (r *rpcRequest) Payload() interface{}        { return r.payload }
func (r *rpcRequest) ContentType() string         { return r.contentType }
func (r *rpcRequest) Service() string             { return r.service }
func (r *rpcRequest) Operation() string           { return r.method }
func (r *rpcRequest) Endpoint() string            { return r.url }
func (r *rpcRequest) Stream() bool                { return r.stream != nil }
