package orm

import (
	"gorm.io/gorm"
)

const Name = "orm"

type Client struct {
	*gorm.DB
}

func (c *Client) Ping() error {
	var _db, err = c.DB.DB()
	if err != nil {
		return err
	}
	return _db.Ping()
}
