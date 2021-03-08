package metric

import (
	"github.com/pubgo/golug/consts"
	"github.com/pubgo/golug/types"
	"github.com/pubgo/xerror"
)

var reporters types.SMap

func Get(names ...string) Reporter {
	val, ok := reporters.Load(consts.GetDefault(names...))
	if !ok {
		return nil
	}

	return val.(Reporter)
}

func List() (dt map[string]Reporter) {
	xerror.Panic(reporters.Map(&dt))
	return
}

func Register(name string, r Reporter) (err error) {
	defer xerror.RespErr(&err)

	xerror.Assert(name == "" || r == nil || r.Name() == "", "[name,reporter] is null")
	xerror.Assert(reporters.Has(name), "reporter %s already exists", name)
	reporters.Set(name, r)
	return
}
