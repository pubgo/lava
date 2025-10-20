package lava

import (
	"github.com/valyala/fasthttp"
)

type ResponseHeader = fasthttp.ResponseHeader

// Response is the response writer interface
type Response interface {
	// Header returns the response header
	Header() *ResponseHeader

	// Payload returns the response payload, []byte or protobuf message
	Payload() any

	// Stream returns true if the response is a stream
	Stream() bool
}
