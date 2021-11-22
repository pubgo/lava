package restc

import (
	"net/http"

	"github.com/pubgo/lava/pkg/encoding"
	"github.com/pubgo/lava/types"
)

var _ types.Request = (*Request)(nil)

type Request struct {
	*http.Request
	service string
	ct      string
	cdc     encoding.Codec
	data    []byte
}

func (r *Request) Kind() string          { return Name }
func (r *Request) Codec() encoding.Codec { return r.cdc }
func (r *Request) Client() bool          { return true }
func (r *Request) Service() string       { return r.service }
func (r *Request) Method() string        { return r.Request.Method }
func (r *Request) Endpoint() string      { return r.Request.RequestURI }
func (r *Request) ContentType() string   { return r.ct }
func (r *Request) Header() types.Header  { return types.Header(r.Request.Header) }
func (r *Request) Payload() interface{}  { return r.data }
func (r *Request) Read() ([]byte, error) { return r.data, nil }
func (r *Request) Stream() bool          { return false }
