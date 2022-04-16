package grpcc_builder

import (
	"github.com/pubgo/lava/abc"
	"google.golang.org/grpc"
)

var _ abc.Response = (*response)(nil)

type response struct {
	header *abc.ResponseHeader
	stream grpc.ClientStream
	resp   interface{}
}

func (r *response) Stream() bool                { return r.stream != nil }
func (r *response) Header() *abc.ResponseHeader { return r.header }
func (r *response) Payload() interface{}        { return r.resp }
