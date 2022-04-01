package service_inter

import "google.golang.org/grpc/metadata"

// Response is the response writer for un encoded messages
type Response interface {
	Header() metadata.MD
	Payload() interface{}
	Stream() bool
}
