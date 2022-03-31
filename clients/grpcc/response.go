package grpcc

import (
	"github.com/pubgo/lava/service"
	"google.golang.org/grpc"
)

var _ service.Response = (*response)(nil)

type response struct {
	req    *request
	stream grpc.ClientStream
	resp   interface{}
}

func (r *response) Stream() bool           { return r.stream != nil }
func (r *response) Header() service.Header { return r.req.header }
func (r *response) Payload() interface{}   { return r.resp }
