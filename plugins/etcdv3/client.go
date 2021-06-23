package etcdv3

import (
	"go.etcd.io/etcd/client/v3"
)

type Client struct {
	*clientv3.Client
}
