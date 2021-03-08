package etcdv3

import (
	"github.com/pubgo/xlog"
	"github.com/pubgo/xlog/xlog_abc"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc/grpclog"
)

var log xlog.Xlog

func init() {
	xlog.Watch(func(logs xlog_abc.Xlog) {
		log = logs.Named(Name)
		grpclog.NewLoggerV2()
		clientv3.SetLogger()
	})
}
