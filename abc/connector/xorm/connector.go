package xorm

import (
	"context"
	"github.com/pubgo/lug/db"

	"github.com/pubgo/lug/abc/connector"

	"github.com/pubgo/xerror"
	"github.com/valyala/fasttemplate"
	"xorm.io/xorm"
)

var _ connector.Connector = (*Connector)(nil)

type Connector struct {
	ResourceID string `json:"resource-id"`
	Table      string `json:"table"`
	Sql        string `json:"sql"`

	tmpl   *fasttemplate.Template
	source *db.Client `dix:""`
}

func (c *Connector) Read(ctx context.Context, cb func(interface{})) (gErr error) {
	defer xerror.RespErr(&gErr)
	var sqlStr = c.tmpl.ExecuteString(map[string]interface{}{})
	var dbResp, err = c.get().Context(ctx).QueryInterface(sqlStr)
	if err != nil {
		return xerror.Wrap(err)
	}

	for i := range dbResp {
		cb(dbResp[i])
	}

	return nil
}

func (c *Connector) Write(ctx context.Context, data interface{}) (err error) {
	defer xerror.RespErr(&err)
	_ = xerror.PanicErr(c.get().Context(ctx).Table(c.Table).Insert(data))
	return nil
}

func (c *Connector) Build() (err error) {
	defer xerror.RespErr(&err)

	c.tmpl = fasttemplate.New(c.Sql, "{", "}")

	// resource处理
	xerror.Assert(c.source == nil, "resource [%s] not found", Name)
	return nil
}

func (c *Connector) Close() error {
	return nil
}

func (c *Connector) get() *xorm.Engine {
	return c.source.Get()
}
