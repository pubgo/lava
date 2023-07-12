package grpcc

import (
	"google.golang.org/grpc"

	"github.com/pubgo/lava/lava"
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
