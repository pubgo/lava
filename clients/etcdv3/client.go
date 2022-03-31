package etcdv3

import (
	"github.com/pubgo/lava/resource"
	client3 "go.etcd.io/etcd/client/v3"
)

type Client struct {
	resource.Resource
}

func (c *Client) Get() *client3.Client {
	var obj = c.Load()
	defer c.Done()
	return obj
}

func (c *Client) Load() *client3.Client {
	return c.GetRes().(*client3.Client)
}
