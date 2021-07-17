package entry

import "context"

// Response is the response writer for unencoded messages
type Response interface {
	// Codec Encoded writer
	Codec() string
	// WriteHeader Write the header
	WriteHeader(map[string]string)
	// Write write a response directly to the client
	Write([]byte) error
}

// Stream represents a stream established with a client.
// A stream can be bidirectional which is indicated by the request.
// The last error will be left in Error().
// EOF indicates end of the stream.
type Stream interface {
	Context() context.Context
	Request() Request
	Send(interface{}) error
	Recv(interface{}) error
	Error() error
	Close() error
}
