package abc

import (
	"github.com/valyala/fasthttp"
)

type ResponseHeader = fasthttp.ResponseHeader

// Response is the response writer for un encoded messages
type Response interface {
	Header() *ResponseHeader
	Payload() interface{}
	Stream() bool
}
