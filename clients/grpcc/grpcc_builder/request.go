package grpcc_builder

import (
	"google.golang.org/grpc"

	"github.com/pubgo/lava/clients/grpcc/grpcc_config"
	"github.com/pubgo/lava/middleware"
)

var _ middleware.Request = (*request)(nil)

type request struct {
	ct       string
	opts     []grpc.CallOption
	method   string
	service  string
	req      interface{}
	cc       *grpc.ClientConn
	invoker  grpc.UnaryInvoker
	streamer grpc.Streamer
	desc     *grpc.StreamDesc
	header   *middleware.RequestHeader
}

func (r *request) Operation() string                 { return r.method }
func (r *request) Kind() string                      { return grpcc_config.Name }
func (r *request) Client() bool                      { return true }
func (r *request) Service() string                   { return r.service }
func (r *request) Endpoint() string                  { return r.method }
func (r *request) ContentType() string               { return r.ct }
func (r *request) Header() *middleware.RequestHeader { return r.header }
func (r *request) Payload() interface{}              { return r.req }
func (r *request) Stream() bool                      { return r.desc != nil }
