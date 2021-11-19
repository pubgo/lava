package restc

import (
	"github.com/pubgo/lava/pkg/encoding"
	"github.com/pubgo/lava/types"
)

var _ types.Request = (*request)(nil)

type request struct {
	service string
	req     *Request
	ct      string
	cdc     encoding.Codec
}

func (r *request) Kind() string          { return Name }
func (r *request) Codec() encoding.Codec { return r.cdc }
func (r *request) Client() bool          { return true }
func (r *request) Service() string       { return r.service }
func (r *request) Method() string        { return r.req.Method }
func (r *request) Endpoint() string      { return r.req.RequestURI }
func (r *request) ContentType() string   { return r.ct }
func (r *request) Header() types.Header  { return types.Header(r.req.Header) }
func (r *request) Payload() interface{}  { return nil }
func (r *request) Read() ([]byte, error) { return nil, nil }
func (r *request) Stream() bool          { return false }
