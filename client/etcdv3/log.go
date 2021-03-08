package etcdv3

import (
	"github.com/pubgo/xlog"
	"github.com/pubgo/xlog/xlog_abc"
)

var log xlog.Xlog

func init() {
	xlog.Watch(func(logs xlog_abc.Xlog) {
		log = logs.Named(Name)
	})
}
