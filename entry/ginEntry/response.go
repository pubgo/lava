package ginEntry

import (
	"github.com/gin-gonic/gin"
	"github.com/pubgo/lava/types"
)

var _ types.Response = (*httpResponse)(nil)

type httpResponse struct {
	ctx *gin.Context
}

func (h *httpResponse) Stream() bool {
	return false
}

func (h *httpResponse) Header() types.Header {
	return types.Header(h.ctx.Writer.Header())
}

func (h *httpResponse) Body() ([]byte, error) {
	return nil, nil
}

func (h *httpResponse) Payload() interface{} {
	return nil
}
