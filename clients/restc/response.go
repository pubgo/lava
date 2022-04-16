package restc

import (
	"github.com/pubgo/lava/abc"
	"github.com/valyala/fasthttp"
)

var _ abc.Response = (*Response)(nil)

type Response struct {
	resp *fasthttp.Response
}

func (r *Response) Header() *abc.ResponseHeader { return &r.resp.Header }
func (r *Response) Payload() interface{}        { return nil }
func (r *Response) Stream() bool                { return false }
