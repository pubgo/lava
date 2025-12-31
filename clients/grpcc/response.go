package grpcc

import (
	"github.com/pubgo/lava/v2/lava"
	"google.golang.org/grpc"
)

var _ lava.Response = (*response)(nil)

type response struct {
	header *lava.ResponseHeader
	stream grpc.ClientStream
	resp   any
}

func (r *response) Stream() bool                 { return r.stream != nil }
func (r *response) Header() *lava.ResponseHeader { return r.header }
func (r *response) Payload() any                 { return r.resp }
