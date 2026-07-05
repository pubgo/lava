package resty

import (
	"fmt"

	"github.com/gofiber/utils/v2"
	"github.com/pubgo/funk/v2/convert"
	"github.com/valyala/fasthttp"

	"github.com/pubgo/lava/v2/pkg/lava"
	"github.com/pubgo/lava/v2/pkg/httputil"
)

var _ lava.Request = (*requestImpl)(nil)

type requestImpl struct {
	req     *Request
	service string
}

func (r *requestImpl) Operation() string {
	return fmt.Sprintf("%s %s", r.req.cfg.Method, r.req.operation)
}

func (r *requestImpl) Kind() string                { return Name }
func (r *requestImpl) Client() bool                { return true }
func (r *requestImpl) Service() string             { return r.service }
func (r *requestImpl) Endpoint() string            { return utils.UnsafeString(r.req.req.URI().Path()) }
func (r *requestImpl) ContentType() string         { return convert.B2S(r.req.req.Header.ContentType()) }
func (r *requestImpl) Header() lava.RequestHeader { return httputil.WrapRequestHeader(&r.req.req.Header) }
func (r *requestImpl) Payload() any                { return r.req.req.Body() }
func (r *requestImpl) Stream() bool                { return false }

var _ lava.Response = (*responseImpl)(nil)

type responseImpl struct {
	resp *fasthttp.Response
}

func (r *responseImpl) Header() lava.ResponseHeader { return httputil.WrapResponseHeader(&r.resp.Header) }
func (r *responseImpl) Payload() any                 { return r.resp.Body() }
func (r *responseImpl) Stream() bool                 { return false }
