package types

import "github.com/pubgo/lava/pkg/encoding"

// Request is a synchronous request interface
type Request interface {
	// Client server or client
	Client() bool

	// Kind server or client kind
	Kind() string

	// Service name requested
	Service() string

	// Method The action requested
	Method() string

	// Endpoint name requested
	Endpoint() string

	// ContentType Content type provided
	ContentType() string

	// Header of the request
	Header() Header

	// Codec The encoded message
	Codec() encoding.Codec

	// Payload is the decoded value
	Payload() interface{}

	// Read the encode request body
	Read() ([]byte, error)

	// Stream Indicates whether its a stream
	Stream() bool
}
