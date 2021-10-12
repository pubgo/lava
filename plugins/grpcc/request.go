package grpcc

import (
	"net/http"

	"google.golang.org/grpc"

	"github.com/pubgo/lava/types"
)

var _ types.Request = (*request)(nil)

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
	header     http.Header
}

func (r *request) Client() bool {
	return true
}

func (r *request) Service() string {
	return r.service
}

func (r *request) Method() string {
	return r.method
}

func (r *request) Endpoint() string {
	return r.method
}

func (r *request) ContentType() string {
	return r.ct
}

func (r *request) Header() types.Header {
	return r.header
}

func (r *request) Payload() interface{} {
	return r.req
}

func (r *request) Body() ([]byte, error) {
	return nil, nil
}

func (r *request) Codec() string {
	return r.ct
}

func (r *request) Stream() bool {
	return false
}
