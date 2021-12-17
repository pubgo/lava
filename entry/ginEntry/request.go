package ginEntry

import (
	"github.com/gin-gonic/gin"
	"github.com/pubgo/lava/encoding"

	"github.com/pubgo/lava/types"
)

var _ types.Request = (*httpRequest)(nil)

type httpRequest struct {
	data        []byte
	ctx         *gin.Context
	contentType string
	cdc         encoding.Codec
	ct          string
}

func (r *httpRequest) Kind() string          { return Name }
func (r *httpRequest) Codec() encoding.Codec { return r.cdc }
func (r *httpRequest) Client() bool          { return false }
func (r *httpRequest) Header() types.Header  { return types.Header(r.ctx.Request.Header) }
func (r *httpRequest) ContentType() string   { return r.ct }
func (r *httpRequest) Service() string       { return r.ctx.Request.Host }
func (r *httpRequest) Method() string        { return r.ctx.FullPath() }
func (r *httpRequest) Endpoint() string      { return r.ctx.Request.RequestURI }
func (r *httpRequest) Stream() bool          { return false }
func (r *httpRequest) Payload() interface{}  { return r.data }
func (r *httpRequest) Read() ([]byte, error) { return r.data, nil }
