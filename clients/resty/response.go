package resty

import (
	"github.com/valyala/fasthttp"

	"github.com/pubgo/lava/lava"
)

var _ lava.Response = (*responseImpl)(nil)

type responseImpl struct {
	resp *fasthttp.Response
}

func (r *responseImpl) Header() *lava.ResponseHeader { return &r.resp.Header }
func (r *responseImpl) Payload() interface{}         { return r.resp.Body() }
func (r *responseImpl) Stream() bool                 { return false }
