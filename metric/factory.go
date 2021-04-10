package metric

import (
	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/types"
	"github.com/pubgo/xerror"
)

type Factory func(cfg map[string]interface{}) (Reporter, error)

var reporters types.SMap

func Get(names ...string) Factory {
	val, ok := reporters.Load(consts.GetDefault(names...))
	if !ok {
		return nil
	}

	return val.(Factory)
}

func Register(name string, r Factory) (err error) {
	defer xerror.RespErr(&err)

	xerror.Assert(name == "" || r == nil, "[name,reporter] is null")
	xerror.Assert(reporters.Has(name), "reporter %s already exists", name)
	reporters.Set(name, r)
	return
}
