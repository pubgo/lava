package grpcEntry

import (
	"github.com/pubgo/xlog"
	"github.com/pubgo/xlog/xlog_abc"
	"github.com/pubgo/xlog/xlog_grpc"
)

func init() {
	xlog.Watch(func(logs xlog_abc.Xlog) {
		xlog_grpc.Init(logs.Named("grpc"))
	})
}
