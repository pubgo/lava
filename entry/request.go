package entry

// Request is a synchronous request interface
type Request interface {
	// Service name requested
	Service() string
	// Method The action requested
	Method() string
	// Endpoint name requested
	Endpoint() string
	// ContentType Content type provided
	ContentType() string
	// Header of the request
	Header() map[string]string
	// Body is the initial decoded value
	Body() interface{}
	// Read the encode request body
	Read() ([]byte, error)
	// Codec The encoded message stream
	Codec() string
	// Stream Indicates whether its a stream
	Stream() bool
}
