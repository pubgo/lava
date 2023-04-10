package orm

import (
	"database/sql"

	"github.com/pubgo/funk/result"
	"gorm.io/gorm"

	"github.com/pubgo/lava/core/vars"
)

const Name = "orm"

type Client struct {
	*gorm.DB
	TablePrefix string
}

func (c *Client) Ping() error {
	_db, err := c.DB.DB()
	if err != nil {
		return err
	}
	return _db.Ping()
}

func (c *Client) Vars() vars.Value {
	return func() interface{} {
		_db, err := c.DB.DB()
		if err != nil {
			return err.Error()
		} else {
			return _db.Stats()
		}
	}
}

func (c *Client) Close() error {
	db, err := c.DB.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

func (c *Client) Stats() (r result.Result[sql.DBStats]) {
	db, err := c.DB.DB()
	if err != nil {
		return r.WithErr(err)
	}
	return r.WithVal(db.Stats())
}
