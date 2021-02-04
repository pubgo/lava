package etcdv3

import "github.com/coreos/etcd/clientv3"

type Client struct {
	*clientv3.Client
}
