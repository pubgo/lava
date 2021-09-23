package stdout

import (
	"github.com/pubgo/lug/abc/connector"
	"github.com/pubgo/lug/abc/factory"
	"github.com/pubgo/lug/types"

	"github.com/pubgo/xerror"
)

const Name = "stdout"

func init() {
	xerror.Exit(factory.Register(&factory.Factory{
		Name: Name,
		Type: connector.Name,
		Handler: func(cfg types.CfgMap) (_ interface{}, err error) {
			defer xerror.RespErr(&err)
			return new(Connector), nil
		},
	}))
}
