package service_builder

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lava/middleware"
	"github.com/pubgo/x/byteutil"
)

var _ middleware.Request = (*httpRequest)(nil)

type httpRequest struct {
	ctx *fiber.Ctx
}

func (r *httpRequest) Kind() string                      { return "http" }
func (r *httpRequest) Operation() string                 { return r.ctx.Route().Path }
func (r *httpRequest) Client() bool                      { return false }
func (r *httpRequest) Header() *middleware.RequestHeader { return &r.ctx.Request().Header }
func (r *httpRequest) Payload() interface{}              { return r.ctx.Body() }

func (r *httpRequest) ContentType() string {
	return byteutil.ToStr(r.ctx.Request().Header.ContentType())
}

func (r *httpRequest) Service() string  { return r.ctx.OriginalURL() }
func (r *httpRequest) Endpoint() string { return r.ctx.OriginalURL() }
func (r *httpRequest) Stream() bool     { return false }
