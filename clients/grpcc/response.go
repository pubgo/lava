package grpcc

import (
	"github.com/pubgo/lava/lava"
	"google.golang.org/grpc"
)

var _ lava.Response = (*response)(nil)

type response struct {
	header *lava.ResponseHeader
	stream grpc.ClientStream
	resp   interface{}
}

func (r *response) Stream() bool                 { return r.stream != nil }
func (r *response) Header() *lava.ResponseHeader { return r.header }
func (r *response) Payload() interface{}         { return r.resp }
