package resty

import (
	"github.com/pubgo/lava/service"
	"github.com/valyala/fasthttp"
)

var _ service.Response = (*Response)(nil)

type Response struct {
	resp *fasthttp.Response
}

func (r *Response) Header() *service.ResponseHeader { return &r.resp.Header }
func (r *Response) Payload() interface{}            { return nil }
func (r *Response) Stream() bool                    { return false }
