package restEntry

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/x/byteutil"

	"github.com/pubgo/lug/types"
)

var _ types.Response = (*httpResponse)(nil)

type httpResponse struct {
	ctx    *fiber.Ctx
	header types.Header
}

func (h *httpResponse) Write(p []byte) (n int, err error) {
	return h.ctx.Write(p)
}

func (h *httpResponse) Header() types.Header {
	return h.header
}

func (h *httpResponse) Body() ([]byte, error) {
	return h.ctx.Response().Body(), nil
}

func (h *httpResponse) Payload() interface{} {
	return h.ctx.Response().Body()
}

func (h *httpResponse) Codec() string {
	return byteutil.ToStr(h.ctx.Response().Header.ContentType())
}

func (h *httpResponse) Send(i interface{}) error {
	panic("implement me")
}

func (h *httpResponse) Recv(i interface{}) error {
	panic("implement me")
}

func (h *httpResponse) Stream() bool {
	return false
}
