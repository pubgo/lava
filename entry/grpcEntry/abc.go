package grpcEntry

import (
	"github.com/pubgo/lava/entry"

	// grpc log插件加载
	_ "github.com/pubgo/lava/internal/plugins/grpclog"
)

type Entry interface {
	entry.Entry
	Register(handler entry.InitHandler)
}
