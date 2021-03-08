package metric

import (
	"github.com/pubgo/golug/consts"
	"github.com/pubgo/golug/types"
	"github.com/pubgo/xerror"
)

var reporters types.SMap

func List() (dt map[string]Reporter) {
	xerror.Panic(reporters.Map(&dt))
	return
}

func Register(r Reporter) (err error) {
	defer xerror.RespErr(&err)

	xerror.Assert(r == nil || r.String() == "", "[reporter] is null")

	schema := r.String()
	xerror.Assert(reporters.Has(schema), "reporter %s already exists", schema)

	reporters.Set(schema, r)
	return
}

func Get(schemas ...string) Reporter {
	val, ok := reporters.Load(consts.GetDefault(schemas...))
	if !ok {
		return nil
	}

	return val.(Reporter)
}
