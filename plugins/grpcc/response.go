package grpcc

import (
	"encoding/json"

	"github.com/pubgo/lug/types"
	"google.golang.org/grpc"
)

var _ types.Response = (*response)(nil)

type response struct {
	req    *request
	stream grpc.ClientStream
	resp   interface{}
}

func (r *response) Write(p []byte) (n int, err error) {
	panic("implement me")
}

func (r *response) Codec() string {
	return r.req.Codec()
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

func (r *response) Send(i interface{}) error {
	return r.stream.SendMsg(i)
}

func (r *response) Recv(i interface{}) error {
	return r.stream.RecvMsg(i)
}

func (r *response) Stream() bool {
	return r.stream != nil
}
