package etcdv3

import (
	"go.etcd.io/etcd/clientv3"
)

type Client struct {
	*clientv3.Client
}
