package grpcEntry

import (
	"github.com/pubgo/lava/entry"
)

type options struct{}
type Opt func(opts *options)
type Entry interface {
	entry.Entry
	Register(handler interface{}, opts ...Opt)
	//UnaryInterceptor(interceptors ...grpc.UnaryServerInterceptor)
	//StreamInterceptor(interceptors ...grpc.StreamServerInterceptor)
}


