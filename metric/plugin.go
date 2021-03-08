package metric

import (
	"github.com/pubgo/golug/config"
	"github.com/pubgo/golug/internal/golug_run"
	"github.com/pubgo/golug/plugin"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(ent interface{}) {
			reporters := List()
			if len(reporters) == 0 {
				xlog.Warn("reporter list is zero")
				return
			}

			config.Decode(Name, &cfg)

			reporter := Get(cfg)
			xerror.Assert(reporter == nil, "reporter %s is null", cfg)
			xerror.Panic(reporter.Start())

			// 停止服务之后, 关闭配置的监控
			golug_run.AfterStop(func() { xerror.Panic(reporter.Stop()) })
		},
	})
}
