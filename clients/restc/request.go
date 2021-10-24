package restc

import (
	"github.com/pubgo/lava/types"
	"github.com/pubgo/x/byteutil"
)

var _ types.Request = (*request)(nil)

type request struct {
	header types.Header
	req    *Request
}

func (r *request) Client() bool {
	return true
}

func (r *request) Service() string {
	return ""
}

func (r *request) Method() string {
	return byteutil.ToStr(r.req.Header.Method())
}

func (r *request) Endpoint() string {
	return r.req.URI().String()
}

func (r *request) ContentType() string {
	return byteutil.ToStr(r.req.Header.ContentType())
}

func (r *request) Header() types.Header {
	return r.header
}

func (r *request) Payload() interface{} {
	return nil
}

func (r *request) Body() ([]byte, error) {
	return r.req.Body(), nil
}

func (r *request) Codec() string {
	return byteutil.ToStr(r.req.Header.ContentType())
}

func (r *request) Stream() bool {
	return false
}
