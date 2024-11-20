package https

import (
	fiber "github.com/gofiber/fiber/v3"

	"github.com/pubgo/lava/lava"
)

var _ lava.Request = (*httpRequest)(nil)

type httpRequest struct {
	ctx fiber.Ctx
}

func (r *httpRequest) Kind() string                { return "http" }
func (r *httpRequest) Operation() string           { return r.ctx.Route().Path }
func (r *httpRequest) Client() bool                { return false }
func (r *httpRequest) Header() *lava.RequestHeader { return &r.ctx.Request().Header }
func (r *httpRequest) Payload() interface{}        { return r.ctx.Body() }

func (r *httpRequest) ContentType() string {
	return string(r.ctx.Request().Header.ContentType())
}

func (r *httpRequest) Service() string  { return r.ctx.OriginalURL() }
func (r *httpRequest) Endpoint() string { return string(r.ctx.Request().RequestURI()) }
func (r *httpRequest) Stream() bool     { return false }
