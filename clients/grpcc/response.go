package grpcc

import (
	"github.com/pubgo/lava/service"
	"github.com/pubgo/lava/service/service_type"

	"google.golang.org/grpc"
)

var _ service_type.Response = (*response)(nil)

type response struct {
	req    *request
	stream grpc.ClientStream
	resp   interface{}
}

func (r *response) Stream() bool           { return r.stream != nil }
func (r *response) Header() service.Header { return r.req.header }
func (r *response) Payload() interface{}   { return r.resp }
