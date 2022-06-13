package orm

import (
	"context"
	"database/sql"
	"errors"

	"github.com/pubgo/xerror"
	"gorm.io/gorm"

	"github.com/pubgo/lava/vars"
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

func (c *Client) InitTable(tb interface{}) error {
	if !c.Migrator().HasTable(tb) {
		return c.AutoMigrate(tb)
	}
	return nil
}

func (c *Client) Upsert(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	var db = c.DB.WithContext(ctx)

	var count int64
	xerror.Panic(db.Model(dest).Where(query, args...).Count(&count).Error)
	if count == 0 {
		return db.Save(dest).Error
	} else {
		return db.Where(query, args...).Updates(dest).Error
	}
}

func (c *Client) Close() error {
	var db, err = c.DB.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

func (c *Client) Stats() (sql.DBStats, error) {
	var db, err = c.DB.DB()
	if err != nil {
		return sql.DBStats{}, err
	}
	return db.Stats(), nil
}

func ErrNotFound(err error) bool {
	if err == gorm.ErrRecordNotFound {
		return true
	}

	return errors.Is(err, gorm.ErrRecordNotFound)
}
