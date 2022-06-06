package restc

import (
	"github.com/pubgo/lava/core/middleware"
	"github.com/valyala/fasthttp"
)

var _ middleware.Response = (*Response)(nil)

type Response struct {
	resp *fasthttp.Response
}

func (r *Response) Header() *middleware.ResponseHeader { return &r.resp.Header }
func (r *Response) Payload() interface{}               { return nil }
func (r *Response) Stream() bool                       { return false }
