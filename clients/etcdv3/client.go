package etcdv3

import (
	client3 "go.etcd.io/etcd/client/v3"

	"github.com/pubgo/lava/pkg/utils"
	"github.com/pubgo/lava/resource"
)

// Get 获取etcd client
func Get(names ...string) *Client {
	var c = resource.Get(Name, utils.GetDefault(names...))
	if c != nil {
		return c.(*Client)
	}
	return nil
}

type Client struct {
	resource.Resource
}

func (c *Client) Get() *client3.Client {
	var obj, r = c.Load()
	defer r.Release()
	return obj
}

func (c *Client) Load() (*client3.Client, resource.Release) {
	var obj, cancel = c.LoadObj()
	return obj.(*client3.Client), cancel
}
