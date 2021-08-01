package types

// Response is the response writer for un encoded messages
type Response interface {
	Header() Header
	Write(p []byte) (n int, err error)
	Codec() string
	Body() ([]byte, error)
	Payload() interface{}
	Send(interface{}) error
	Recv(interface{}) error
	Stream() bool
}
