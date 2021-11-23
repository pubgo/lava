package etcdv3

import (
	"github.com/pubgo/lava/pkg/lavax"
	"github.com/pubgo/lava/resource"
	client3 "go.etcd.io/etcd/client/v3"
)

// Get 获取etcd client
func Get(names ...string) *Client {
	var c = resource.Get(Name, lavax.GetDefault(names...))
	if c != nil {
		return c.(*Client)
	}
	return nil
}

var _ resource.Resource = (*Client)(nil)

type Client struct {
	*client3.Client
}

func (c *Client) UpdateResObj(val interface{}) { c.Client = val.(*Client).Client }
func (c *Client) Kind() string                 { return Name }
