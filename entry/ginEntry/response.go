package ginEntry

import (
	"github.com/gin-gonic/gin"
	"github.com/pubgo/lava/service"
	"github.com/pubgo/lava/service/service_type"
)

var _ service_type.Response = (*httpResponse)(nil)

type httpResponse struct {
	ctx *gin.Context
}

func (h *httpResponse) Stream() bool { return false }
func (h *httpResponse) Header() service.Header {
	return service.Header(h.ctx.Writer.Header())
}
func (h *httpResponse) Payload() interface{} { return nil }
