package restc

import (
	"net/http"
	"net/http/httptrace"

	"github.com/pubgo/lava/types"
)

var _ types.Response = (*Response)(nil)

type Response struct {
	resp *http.Response
}

func (r *Response) Header() types.Header     { return types.Header(r.resp.Header) }
func (r *Response) Response() *http.Response { return r.resp }
func (r *Response) Payload() interface{}     { return nil }
func (r *Response) Stream() bool             { return false }
func (r *Response) TraceInfo() *httptrace.ClientTrace {
	return httptrace.ContextClientTrace(r.resp.Request.Context())
}
