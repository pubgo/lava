package metric

import (
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/vars"
	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
	"github.com/uber-go/tally"
)

func init() {
	vars.Register(Name+"_factory", func() interface{} {
		var dt = make(map[string]string)
		xerror.Panic(factories.Each(func(name string, r Factory) {
			dt[name] = stack.Func(r)
		}))
		return dt
	})

	vars.Register(Name+"_capabilities", func() interface{} {
		var c = GetGlobal().Capabilities()
		return typex.M{
			"reporting": c.Reporting(),
			"tagging":   c.Tagging(),
		}
	})

	vars.Register(Name+"_snapshot", func() interface{} {
		if c, ok := GetGlobal().(tally.TestScope); ok {
			// TODO 数据序列化处理
			return c.Snapshot()
		}
		return nil
	})
}
