package grpcc

import (
	"google.golang.org/grpc"

	"github.com/pubgo/lava/core/middleware"
)

var _ middleware.Response = (*response)(nil)

type response struct {
	header *middleware.ResponseHeader
	stream grpc.ClientStream
	resp   interface{}
}

func (r *response) Stream() bool                       { return r.stream != nil }
func (r *response) Header() *middleware.ResponseHeader { return r.header }
func (r *response) Payload() interface{}               { return r.resp }
