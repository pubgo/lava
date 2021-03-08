package etcdv3

import (
	"github.com/pubgo/xlog"
	"go.etcd.io/etcd/clientv3"
)

type Client struct {
	log xlog.Xlog
	*clientv3.Client
}
