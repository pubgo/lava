package db

import (
	"xorm.io/xorm"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/internal/resource"
)

func Get(names ...string) *xorm.Engine {
	c := resource.Get(Name, consts.GetDefault())
	if c == nil {
		return nil
	}

	return c.(*Client).db
}

func GetCallback(name string, cb func(*xorm.Engine)) {
	c := resource.Get(Name, consts.GetDefault())
	if c == nil {
		return
	}

	cb(c.(*Client).db)
}

var _ resource.Resource = (*Client)(nil)

type Client struct {
	db *xorm.Engine
}

func (c *Client) Close() error      { return c.db.Close() }
func (c *Client) Get() *xorm.Engine { return c.db }
