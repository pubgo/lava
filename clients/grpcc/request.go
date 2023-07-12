package grpcc

import (
	"google.golang.org/grpc"

	"github.com/pubgo/lava/lava"
)

var _ lava.Request = (*request)(nil)

type request struct {
	resp     grpc.ClientStream
	reply    interface{}
	ct       string
	opts     []grpc.CallOption
	method   string
	service  string
	req      interface{}
	cc       *grpc.ClientConn
	invoker  grpc.UnaryInvoker
	streamer grpc.Streamer
	desc     *grpc.StreamDesc
	header   *lava.RequestHeader
}

func (r *request) Operation() string           { return r.method }
func (r *request) Kind() string                { return Name }
func (r *request) Client() bool                { return true }
func (r *request) Service() string             { return r.service }
func (r *request) Endpoint() string            { return r.method }
func (r *request) ContentType() string         { return r.ct }
func (r *request) Header() *lava.RequestHeader { return r.header }
func (r *request) Payload() interface{}        { return r.req }
func (r *request) Stream() bool                { return r.desc != nil }
