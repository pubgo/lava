package orm

import (
	"database/sql"

	"github.com/pubgo/funk/result"
	"gorm.io/gorm"

	"github.com/pubgo/lava/vars"
)

const Name = "orm"

type Client struct {
	*gorm.DB
}

func (c *Client) Ping() result.Error {
	var _db, err = c.DB.DB()
	if err != nil {
		return result.WithErr(err)
	}
	return result.WithErr(_db.Ping())
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

func (c *Client) InitTable(tb interface{}) result.Error {
	if !c.Migrator().HasTable(tb) {
		return result.WithErr(c.AutoMigrate(tb))
	}
	return result.Error{}
}

func (c *Client) Close() result.Error {
	var db, err = c.DB.DB()
	if err != nil {
		return result.WithErr(err)
	}
	return result.WithErr(db.Close())
}

func (c *Client) Stats() result.Result[sql.DBStats] {
	var db, err = c.DB.DB()
	if err != nil {
		return result.Wrap(sql.DBStats{}, err)
	}
	return result.OK(db.Stats())
}
