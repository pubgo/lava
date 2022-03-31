package restc

import (
	"net/http"
	"net/http/httptrace"

	"github.com/pubgo/lava/service"
	"github.com/pubgo/lava/service/service_type"
)

var _ service_type.Response = (*Response)(nil)

type Response struct {
	resp *http.Response
}

func (r *Response) Header() service.Header   { return service.Header(r.resp.Header) }
func (r *Response) Response() *http.Response { return r.resp }
func (r *Response) Payload() interface{}     { return nil }
func (r *Response) Stream() bool             { return false }
func (r *Response) TraceInfo() *httptrace.ClientTrace {
	return httptrace.ContextClientTrace(r.resp.Request.Context())
}
