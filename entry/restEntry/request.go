package restEntry

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/x/byteutil"

	"github.com/pubgo/lug/encoding"
	"github.com/pubgo/lug/types"
)

var _ types.Request = (*httpRequest)(nil)

type httpRequest struct {
	ctx    *fiber.Ctx
	header types.Header
}

func (r *httpRequest) Client() bool {
	return false
}

func (r *httpRequest) Header() types.Header {
	return r.header
}

func (r *httpRequest) Payload() interface{} {
	return r.ctx.Body()
}

func (r *httpRequest) Body() ([]byte, error) {
	return r.ctx.Body(), nil
}

func (r *httpRequest) ContentType() string {
	return byteutil.ToStr(r.ctx.Request().Header.ContentType())
}

func (r *httpRequest) Service() string {
	return r.ctx.OriginalURL()
}

func (r *httpRequest) Method() string {
	return r.ctx.Method()
}

func (r *httpRequest) Endpoint() string {
	return r.ctx.OriginalURL()
}

func (r *httpRequest) Codec() string {
	return encoding.Mapping[r.ContentType()]
}

func (r *httpRequest) Stream() bool {
	return false
}
