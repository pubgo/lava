package resty

import (
	"github.com/gofiber/utils"
	"github.com/valyala/fasthttp"

	"github.com/pubgo/lava/lava"
)

var _ lava.Request = (*requestImpl)(nil)

type requestImpl struct {
	req     *fasthttp.Request
	service string
	ct      string
	data    []byte
}

func (r *requestImpl) Operation() string           { return utils.UnsafeString(r.req.Header.Method()) }
func (r *requestImpl) Kind() string                { return Name }
func (r *requestImpl) Client() bool                { return true }
func (r *requestImpl) Service() string             { return r.service }
func (r *requestImpl) Endpoint() string            { return utils.UnsafeString(r.req.URI().Path()) }
func (r *requestImpl) ContentType() string         { return r.ct }
func (r *requestImpl) Header() *lava.RequestHeader { return &r.req.Header }
func (r *requestImpl) Payload() interface{}        { return r.data }
func (r *requestImpl) Stream() bool                { return false }

var _ lava.Response = (*responseImpl)(nil)

type responseImpl struct {
	resp *fasthttp.Response
}

func (r *responseImpl) Header() *lava.ResponseHeader { return &r.resp.Header }
func (r *responseImpl) Payload() interface{}         { return r.resp.Body() }
func (r *responseImpl) Stream() bool                 { return false }