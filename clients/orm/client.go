package orm

import (
	"gorm.io/gorm"

	"github.com/pubgo/lava/resource"
)

const Name = "gorm"

var _ resource.Resource = (*Client)(nil)

type Client struct {
	*gorm.DB
}

func (c *Client) Ping() error {
	var db, err = c.DB.DB()
	if err != nil {
		return err
	}
	return db.Ping()
}

func (c *Client) Close() error {
	var db, err = c.DB.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

func (c Client) UpdateResObj(val interface{}) { c.DB = val.(*Client).DB }
func (c Client) Kind() string                 { return Name }
