package restc

import (
	"github.com/pubgo/lug/types"
	"github.com/valyala/fasthttp"
)

var _ types.Response = (*response)(nil)

type response struct {
	resp *fasthttp.Response
}

func (r *response) Write(p []byte) (n int, err error) {
	return 0, err
}

func (r *response) Codec() string {
	return ""
}

func (r *response) Header() types.Header {
	panic("implement me")
}

func (r *response) Body() ([]byte, error) {
	return r.resp.Body(), nil
}

func (r *response) Payload() interface{} {
	return r.resp.Body()
}

func (r *response) Send(i interface{}) error {
	panic("implement me")
}

func (r *response) Recv(i interface{}) error {
	panic("implement me")
}

func (r *response) Stream() bool {
	return false
}
