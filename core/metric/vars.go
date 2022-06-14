package metric

import (
	"github.com/pubgo/lava/internal/pkg/typex"
	"github.com/pubgo/lava/vars"
	"github.com/uber-go/tally"
)

func registerVars(m Metric) {
	vars.Register(Name+"_capabilities", func() interface{} {
		var c = m.Capabilities()
		return typex.M{
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
