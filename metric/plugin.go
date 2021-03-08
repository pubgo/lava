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
			registries := List()
			if len(registries) == 0 {
				xlog.Warn("reporter list is zero")
				return
			}

			// 解析registry配置
			config.Decode(Name, &cfg)

			for k, v := range cfg {
				if cfg.Driver != k {
					continue
				}

				xerror.PanicF(v.Init(cfg), "registry %s init error", k)

				// 注册中心只有一个, 所以可以使用Default, 否着需要使用Get(driver)
				Default = v
			}
		},
	})

	// 服务启动后, 启动配置监控
	golug_run.AfterStart(func() {
		for _, w := range List() {
			xerror.ExitF(w.Start(), "watcher %s start error", w.Name())
		}
	})

	// 停止服务之后, 关闭配置的监控
	golug_run.BeforeStop(func() {
		var err error
		for _, w := range List() {
			err = xerror.Append(err, xerror.WrapF(w.Close(), "watcher %s close error", w.Name()))
		}
		xerror.Panic(err)
	})
}
