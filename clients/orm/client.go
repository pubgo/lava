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

type Client struct {
	resource.Resource
}

func (c *Client) Ping() error {
	var db, release = c.Load()
	defer release.Release()

	var _db, err = db.DB()
	if err != nil {
		return err
	}
	return _db.Ping()
}

func (c *Client) Load() (*gorm.DB, resource.Release) {
	var r, cancel = c.Resource.LoadObj()
	return r.(*wrapper).DB, cancel
}

func (c *Client) get() *gorm.DB {
	var r, cancel = c.Resource.LoadObj()
	defer cancel.Release()
	return r.(*wrapper).DB
}
