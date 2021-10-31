package grpcc

import (
	"encoding/json"

	"github.com/pubgo/lava/types"
	"google.golang.org/grpc"
)

var _ types.Response = (*response)(nil)

type response struct {
	req    *request
	stream grpc.ClientStream
	resp   interface{}
}

func (r *response) Stream() bool {
	return r.stream != nil
}

func (r *response) Header() types.Header {
	return r.req.header
}

func (r *response) Body() ([]byte, error) {
	return json.Marshal(r.resp)
}

func (r *response) Payload() interface{} {
	return r.resp
}
