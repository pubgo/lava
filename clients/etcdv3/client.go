package etcdv3

import (
	"github.com/pubgo/lava/pkg/lavax"
	resource2 "github.com/pubgo/lava/resource"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Get 获取etcd client
func Get(names ...string) *Client {
	var c = resource2.Get(Name, lavax.GetDefault(names...))
	if c != nil {
		return c.(*Client)
	}
	return nil
}

// Delete 删除etcd client
func Delete(name string) {
	resource2.Remove(Name, name)
}

var _ resource2.Resource = (*Client)(nil)

type Client struct {
	*clientv3.Client
}

func (c *Client) UpdateResObj(val interface{}) { c.Client = val.(*Client).Client }
func (c *Client) Kind() string                 { return Name }
