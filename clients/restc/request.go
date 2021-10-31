package restc

import (
	"github.com/pubgo/lava/pkg/encoding"
	"github.com/pubgo/lava/types"
)

var _ types.Request = (*request)(nil)

type request struct {
	header types.Header
	req    *Request
}

func (r *request) Kind() string {
	return Name
}

func (r *request) Codec() encoding.Codec {
	return encoding.Get(encoding.Mapping[r.ContentType()])
}

func (r *request) Client() bool {
	return true
}

func (r *request) Service() string {
	return ""
}

func (r *request) Method() string {
	return r.req.Method
}

func (r *request) Endpoint() string {
	return r.req.RequestURI
}

func (r *request) ContentType() string {
	return r.ContentType()
}

func (r *request) Header() types.Header {
	return r.header
}

func (r *request) Payload() interface{} {
	return nil
}

func (r *request) Body() ([]byte, error) {
	return nil, nil
}

func (r *request) Stream() bool {
	return false
}
