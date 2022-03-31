package orm

import (
	"github.com/pubgo/lava/resource"
	"io"

	"github.com/pubgo/xerror"
	"gorm.io/gorm"
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
	var db = c.get()
	var _db, err = db.DB()
	if err != nil {
		return err
	}
	return _db.Ping()
}

func (c *Client) Load() *gorm.DB {
	var r = c.Resource.GetRes()
	return r.(*wrapper).DB
}

func (c *Client) get() *gorm.DB {
	var r = c.Resource.GetRes()
	defer c.Done()
	return r.(*wrapper).DB
}
