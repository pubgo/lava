package registry

import (
	"github.com/pubgo/golug/config"
	"github.com/pubgo/golug/internal/golug_run"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

func init() {
	golug_run.BeforeStart(func() {
		registries := List()
		if len(registries) == 0 {
			xlog.Warn("registry list is zero")
			return
		}

		// 解析registry配置
		config.Decode(Name, &cfg)

		for k, v := range registries {
			if cfg.Driver != k {
				continue
			}

			xerror.PanicF(v.Init(cfg), "registry %s init error", k)

			// 注册中心只有一个, 所以可以使用Default, 否着需要使用Get(driver)
			Default = v
		}
	})
}
