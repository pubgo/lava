package rest

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lug/encoding"
	"github.com/pubgo/lug/types"
	"github.com/pubgo/x/byteutil"
)

var _ types.Request = (*httpRequest)(nil)

type httpRequest struct {
	req *fiber.Ctx
}

func (r *httpRequest) Header() types.Header {
	return convertHeader(&r.req.Request().Header)
}

func (r *httpRequest) Payload() interface{} {
	return r.req.Body()
}

func (r *httpRequest) Body() ([]byte, error) {
	return r.req.Body(), nil
}

func (r *httpRequest) ContentType() string {
	return byteutil.ToStr(r.req.Request().Header.ContentType())
}

func (r *httpRequest) Service() string {
	return r.req.OriginalURL()
}

func (r *httpRequest) Method() string {
	return r.req.Method()
}

func (r *httpRequest) Endpoint() string {
	return r.req.OriginalURL()
}

func (r *httpRequest) Codec() string {
	return encoding.Mapping[r.ContentType()]
}

func (r *httpRequest) Stream() bool {
	return false
}
