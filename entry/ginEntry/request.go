package ginEntry

import (
	"github.com/gin-gonic/gin"

	"github.com/pubgo/lava/pkg/encoding"
	"github.com/pubgo/lava/types"
)

var _ types.Request = (*httpRequest)(nil)

type httpRequest struct {
	ctx    *gin.Context
	header types.Header
}

func (r *httpRequest) Client() bool {
	return false
}

func (r *httpRequest) Header() types.Header {
	return r.header
}

func (r *httpRequest) Payload() interface{} {
	return nil
}

func (r *httpRequest) Body() ([]byte, error) {
	return nil, nil
}

func (r *httpRequest) ContentType() string { return r.ctx.ContentType() }

func (r *httpRequest) Service() string {
	return r.ctx.Request.Host
}

func (r *httpRequest) Method() string {
	return r.ctx.Request.Method
}

func (r *httpRequest) Endpoint() string {
	return r.ctx.Request.RequestURI
}

func (r *httpRequest) Codec() string {
	return encoding.Mapping[r.ContentType()]
}

func (r *httpRequest) Stream() bool {
	return false
}
