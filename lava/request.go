package lava

import (
	"github.com/valyala/fasthttp"
)

type RequestKind = string

const (
	RequestKindHttp RequestKind = "http"
	RequestKindGrpc RequestKind = "grpc"
)

type RequestHeader = fasthttp.RequestHeader

// Request is a synchronous request interface
type Request interface {
	// Client server or client
	Client() bool

	// Kind [http|grpc...]
	Kind() RequestKind

	// Stream Indicates whether it's a stream
	Stream() bool

	// Service name requested, grpc service name
	Service() string

	// Operation requested, grpc method name
	Operation() string

	// Endpoint requested, http router path
	Endpoint() string

	// ContentType Content type provided, application/json, application/grpc
	ContentType() string

	// Header of the request
	Header() *RequestHeader

	// Payload is the decoded value, []byte or proto message
	Payload() any
}
