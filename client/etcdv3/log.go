package etcdv3

import (
	"github.com/pubgo/xlog"
)

var log xlog.Xlog

func init() {
	xlog.Watch(func(logs xlog.Xlog) {
		log = logs.Named(Name)
	})
}
