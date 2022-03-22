package service_type

import (
	"google.golang.org/grpc/metadata"
)

// Request is a synchronous request interface
type Request interface {
	// Client server or client
	Client() bool

	// Kind [http|grpc...]
	Kind() string

	// Stream Indicates whether it's a stream
	Stream() bool

	// Service name requested
	Service() string

	// Operation requested
	Operation() string

	// Endpoint requested
	Endpoint() string

	// ContentType Content type provided
	ContentType() string

	// Header of the request
	Header() metadata.MD

	// Payload is the decoded value
	Payload() interface{}
}
