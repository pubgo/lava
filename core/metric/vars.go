package metric

import (
	"github.com/pubgo/funk/typex"
	"github.com/pubgo/lava/core/vars"
	"github.com/uber-go/tally/v4"
)

func registerVars(m Metric) {
	vars.Register(Name+"_capabilities", func() interface{} {
		c := m.Capabilities()
		return typex.Ctx{
			"reporting": c.Reporting(),
			"tagging":   c.Tagging(),
		}
	})

	vars.Register(Name+"_snapshot", func() interface{} {
		if c, ok := m.(tally.TestScope); ok {
			// TODO 数据序列化处理
			return c.Snapshot()
		}
		return nil
	})
}
