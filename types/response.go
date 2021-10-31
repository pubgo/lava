package types

// Response is the response writer for un encoded messages
type Response interface {
	Header() Header
	Body() ([]byte, error)
	Payload() interface{}
	Stream() bool
}
