package metricbuilder

import (
	"github.com/pubgo/funk/typex"
	"github.com/pubgo/funk/vars"
	"github.com/pubgo/lava/core/metrics"
	"github.com/uber-go/tally/v4"
)

func registerVars(m metrics.Metric) {
	vars.Register(metrics.Name+"_capabilities", func() interface{} {
		c := m.Capabilities()
		return typex.Ctx{
			"reporting": c.Reporting(),
			"tagging":   c.Tagging(),
		}
	})

	vars.Register(metrics.Name+"_snapshot", func() interface{} {
		if c, ok := m.(tally.TestScope); ok {
			// TODO 数据序列化处理
			return c.Snapshot()
		}
		return nil
	})
}
