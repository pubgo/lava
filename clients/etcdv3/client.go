package etcdv3

import (
	"io"

	client3 "go.etcd.io/etcd/client/v3"

	"github.com/pubgo/lava/pkg/lavax"
	"github.com/pubgo/lava/resource"
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
	v *client3.Client
}

func (c *Client) Unwrap() io.Closer               { return c.v }
func (c *Client) UpdateObj(val resource.Resource) { c.v = val.(*Client).v }
func (c *Client) Kind() string                    { return Name }
func (c *Client) Get() *client3.Client            { return c.v }
