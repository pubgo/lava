package etcdv3

import (
	client3 "go.etcd.io/etcd/client/v3"
)

type Client struct {
	*client3.Client
}
