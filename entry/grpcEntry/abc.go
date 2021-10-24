package grpcEntry

import (
	"github.com/pubgo/lava/entry"

	// grpc log插件加载
	_ "github.com/pubgo/lava/internal/plugins/grpclog"
)

type options struct{}
type Opt func(opts *options)
type Entry interface {
	entry.Entry
	Register(handler interface{}, opts ...Opt)
}
