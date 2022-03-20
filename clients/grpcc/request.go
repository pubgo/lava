package grpcc

import (
	"github.com/pubgo/lava/service"
	"github.com/pubgo/lava/service/service_type"

	"google.golang.org/grpc"
)

var _ service_type.Request = (*request)(nil)

type request struct {
	ct         string
	opts       []grpc.CallOption
	method     string
	service    string
	req, reply interface{}
	cc         *grpc.ClientConn
	invoker    grpc.UnaryInvoker
	streamer   grpc.Streamer
	desc       *grpc.StreamDesc
	header     service.Header
}

func (r *request) Operation() string      { return r.method }
func (r *request) Kind() string           { return Name }
func (r *request) Client() bool           { return true }
func (r *request) Service() string        { return r.service }
func (r *request) Endpoint() string       { return r.method }
func (r *request) ContentType() string    { return r.ct }
func (r *request) Header() service.Header { return r.header }
func (r *request) Payload() interface{}   { return r.req }
func (r *request) Stream() bool           { return r.desc != nil }
