package etcdv3

import (
	"time"

	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
)

const Name = "etcdv3"
const DefaultTimeout = time.Second * 2

var DefaultCfg = clientv3.Config{
	DialTimeout: DefaultTimeout,
	DialOptions: []grpc.DialOption{grpc.WithBlock()},
}
