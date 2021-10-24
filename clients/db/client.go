package db

import (
	"xorm.io/xorm"

	"github.com/pubgo/lava/pkg/lavax"
	"github.com/pubgo/lava/resource"
)

func Get(names ...string) *Client {
	c := resource.Get(Name, lavax.GetDefault(names...))
	if c == nil {
		return nil
	}

	return c.(*Client)
}

func GetCallback(name string, cb func(*xorm.Engine)) {
	c := resource.Get(Name, lavax.GetDefault())
	if c == nil {
		return
	}

	cb(c.(*Client).Engine)
}

var _ resource.Resource = (*Client)(nil)

type Client struct {
	*xorm.Engine
}

func (c *Client) UpdateResObj(val interface{}) { c.Engine = val.(*Client).Engine }
func (c *Client) Kind() string                 { return Name }
