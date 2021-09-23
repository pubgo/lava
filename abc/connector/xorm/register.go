package xorm

import (
	"github.com/pubgo/lug/abc/connector"
	"github.com/pubgo/lug/abc/factory"
	"github.com/pubgo/lug/types"

	"github.com/pubgo/xerror"
)

const Name = "xorm"

func init() {
	xerror.Exit(factory.Register(&factory.Factory{
		Name:       Name,
		Type:       connector.Name,
		Descriptor: "",
		Handler: func(cfg types.CfgMap) (_ interface{}, err error) {
			defer xerror.RespErr(&err)
			var conn = new(Connector)
			xerror.Panic(cfg.Decode(conn))
			xerror.Panic(conn.Build())
			return conn, nil
		},
	}))
}
