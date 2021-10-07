package etcdv3

import (
	"github.com/pubgo/lug/internal/resource"
	"go.etcd.io/etcd/client/v3"
)

var _ resource.Resource = (*Client)(nil)

type Client struct {
	*clientv3.Client
}
