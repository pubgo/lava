package service_type

import "github.com/pubgo/lava/service"

// Response is the response writer for un encoded messages
type Response interface {
	Header() service.Header
	Payload() interface{}
	Stream() bool
}
