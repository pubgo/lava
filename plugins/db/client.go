package db

import (
	"github.com/pubgo/lava/pkg/lavax"
	resource2 "github.com/pubgo/lava/resource"
	"xorm.io/xorm"
)

func Get(names ...string) *xorm.Engine {
	c := resource2.Get(Name, lavax.GetDefault())
	if c == nil {
		return nil
	}

	return c.(*Client).Engine
}

func GetCallback(name string, cb func(*xorm.Engine)) {
	c := resource2.Get(Name, lavax.GetDefault())
	if c == nil {
		return
	}

	cb(c.(*Client).Engine)
}

var _ resource2.Resource = (*Client)(nil)

type Client struct {
	*xorm.Engine
}

func (c *Client) UpdateResObj(val interface{}) { c.Engine = val.(*Client).Engine }
func (c *Client) Kind() string                 { return Name }
