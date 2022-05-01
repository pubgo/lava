package metric_builder

import (
	"github.com/pubgo/lava/core/metric"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/vars"

	"github.com/uber-go/tally"
)

func init() {
	vars.Register(metric.Name+"_capabilities", func() interface{} {
		var c = GetGlobal().Capabilities()
		return typex.M{
			"reporting": c.Reporting(),
			"tagging":   c.Tagging(),
		}
	})

	vars.Register(metric.Name+"_snapshot", func() interface{} {
		if c, ok := GetGlobal().(tally.TestScope); ok {
			// TODO 数据序列化处理
			return c.Snapshot()
		}
		return nil
	})
}
