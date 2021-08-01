package grpc

import (
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/plugin"

	"google.golang.org/grpc"
)

type options struct{}
type Opt func(opts *options)
type Entry interface {
	entry.Entry
	Plugin(plugins ...plugin.Plugin)
	Register(handler interface{}, opts ...Opt)
	UnaryInterceptor(interceptors ...grpc.UnaryServerInterceptor)
	StreamInterceptor(interceptors ...grpc.StreamServerInterceptor)
}
