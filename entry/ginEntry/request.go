package ginEntry

import (
	"github.com/gin-gonic/gin"
	"github.com/pubgo/lava/service"
	"github.com/pubgo/lava/service/service_type"
)

var _ service_type.Request = (*httpRequest)(nil)

type httpRequest struct {
	data        []byte
	ctx         *gin.Context
	contentType string
	ct          string
}

func (r *httpRequest) Operation() string      { return r.ctx.FullPath() }
func (r *httpRequest) Kind() string           { return Name }
func (r *httpRequest) Client() bool           { return false }
func (r *httpRequest) Header() service.Header { return service.Header(r.ctx.Request.Header) }
func (r *httpRequest) ContentType() string    { return r.ct }
func (r *httpRequest) Service() string        { return r.ctx.Request.Host }
func (r *httpRequest) Endpoint() string       { return r.ctx.Request.RequestURI }
func (r *httpRequest) Stream() bool           { return false }
func (r *httpRequest) Payload() interface{}   { return r.data }
