package orm

import (
	"io"

	"github.com/pubgo/xerror"
	"gorm.io/gorm"

	"github.com/pubgo/lava/resource"
)

const Name = "gorm"

var _ io.Closer = (*wrapper)(nil)

type wrapper struct {
	*gorm.DB
}

func (w *wrapper) Close() error {
	var db, err = w.DB.DB()
	xerror.Panic(err)
	return db.Close()
}

var _ resource.Resource = (*Client)(nil)

type Client struct {
	v *wrapper
}

func (c *Client) Ping() error {
	var db, err = c.v.DB.DB()
	if err != nil {
		return err
	}
	return db.Ping()
}

func (c *Client) Unwrap() io.Closer              { return c.v }
func (c *Client) Get() *gorm.DB                  { return c.v.DB }
func (c Client) UpdateObj(val resource.Resource) { c.v = val.(*Client).v }
func (c Client) Kind() string                    { return Name }
