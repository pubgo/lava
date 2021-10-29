package ginEntry

import (
	"github.com/gin-gonic/gin"
	"github.com/pubgo/lava/types"
)

var _ types.Response = (*httpResponse)(nil)

type httpResponse struct {
	ctx *gin.Context
}

func (h *httpResponse) Write(p []byte) (n int, err error) {
	return h.ctx.Writer.Write(p)
}

func (h *httpResponse) Header() types.Header {
	return h.ctx.Writer.Header()
}

func (h *httpResponse) Body() ([]byte, error) {
	return nil, nil
}

func (h *httpResponse) Payload() interface{} {
	return nil
}

func (h *httpResponse) Codec() string {
	return ""
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
