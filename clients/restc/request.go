package restc

import (
	"github.com/pubgo/lava/service"
	"net/http"
)

var _ service.Request = (*Request)(nil)

type Request struct {
	req     *http.Request
	service string
	ct      string
	data    []byte
}

func (r *Request) Operation() string      { return r.req.Method }
func (r *Request) Kind() string           { return Name }
func (r *Request) Client() bool           { return true }
func (r *Request) Service() string        { return r.service }
func (r *Request) Endpoint() string       { return r.req.RequestURI }
func (r *Request) ContentType() string    { return r.ct }
func (r *Request) Header() service.Header { return service.Header(r.req.Header) }
func (r *Request) Payload() interface{}   { return r.data }
func (r *Request) Stream() bool           { return false }
