package resty

import (
	"fmt"
	"github.com/gofiber/utils"
	"github.com/pubgo/funk/convert"
	"github.com/valyala/fasthttp"

	"github.com/pubgo/lava/lava"
)

var _ lava.Request = (*requestImpl)(nil)

type requestImpl struct {
	req       *fasthttp.Request
	operation string
	service   string
}

func (r *requestImpl) Operation() string {
	return fmt.Sprintf("%s %s", r.req.Header.Method(), r.operation)
}

func (r *requestImpl) Kind() string                { return Name }
func (r *requestImpl) Client() bool                { return true }
func (r *requestImpl) Service() string             { return r.service }
func (r *requestImpl) Endpoint() string            { return utils.UnsafeString(r.req.URI().Path()) }
func (r *requestImpl) ContentType() string         { return convert.B2S(r.req.Header.ContentType()) }
func (r *requestImpl) Header() *lava.RequestHeader { return &r.req.Header }
func (r *requestImpl) Payload() interface{}        { return r.req.Body() }
func (r *requestImpl) Stream() bool                { return false }

var _ lava.Response = (*responseImpl)(nil)

type responseImpl struct {
	resp *fasthttp.Response
}

func (r *responseImpl) Header() *lava.ResponseHeader { return &r.resp.Header }
func (r *responseImpl) Payload() interface{}         { return r.resp.Body() }
func (r *responseImpl) Stream() bool                 { return false }
