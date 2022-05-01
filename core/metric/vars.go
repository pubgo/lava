package metric

import (
	"github.com/pubgo/lava/vars"
	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
)

func init() {
	vars.Register(Name+"_factory", func() interface{} {
		var dt = make(map[string]string)
		xerror.Panic(factories.Each(func(name string, r Factory) {
			dt[name] = stack.Func(r)
		}))
		return dt
	})
}
