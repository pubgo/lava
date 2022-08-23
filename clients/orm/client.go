package orm

import (
	"context"
	"database/sql"
	"errors"

	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/result"
	"github.com/pubgo/funk/xerr"
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

func (c *Client) Upsert(ctx context.Context, dest interface{}, query string, args ...interface{}) (gErr result.Error) {
	defer recovery.Recovery(func(err xerr.XErr) { gErr = result.WithErr(err) })

	var db = c.DB.WithContext(ctx)
	var count int64
	if err := db.Model(dest).Where(query, args...).Count(&count).Error; err != nil {
		return result.WithErr(err)
	}

	if count == 0 {
		return result.WithErr(db.Save(dest).Error)
	} else {
		return result.WithErr(db.Where(query, args...).Updates(dest).Error)
	}
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

func ErrNotFound(err error) bool {
	if err == gorm.ErrRecordNotFound {
		return true
	}

	return errors.Is(err, gorm.ErrRecordNotFound)
}
