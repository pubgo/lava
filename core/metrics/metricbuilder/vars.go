package metricbuilder

import (
	"github.com/pubgo/lava/v2/core/metrics"
	"github.com/uber-go/tally/v4"

	"github.com/pubgo/funk/v2/typex"
	"github.com/pubgo/funk/v2/vars"
)

func registerVars(m metrics.Metric) {
	vars.Register(vars.UniqueName(metrics.Name, "capabilities"), func() any {
		c := m.Capabilities()
		return typex.Ctx{
			"reporting": c.Reporting(),
			"tagging":   c.Tagging(),
		}
	})

	vars.Register(vars.UniqueName(metrics.Name, "snapshot"), func() any {
		if c, ok := m.(tally.TestScope); ok {
			// TODO 数据序列化处理
			return c.Snapshot()
		}
		return nil
	})
}
