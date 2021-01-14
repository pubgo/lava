package golug_registry

import (
	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/pkg/golug_utils"
	"github.com/pubgo/xerror"
)

func init() {
	// 配置初始化完毕后, 解析registry配置
	golug_config.On(func(_ *golug_config.Config) {
		golug_utils.Mergo(&cfg, GetDefaultCfg())
		golug_config.Decode(Name, &cfg)
	})

	xerror.Panic(dix_run.WithBeforeStart(func(ctx *dix_run.BeforeStartCtx) {
		for k, v := range List() {
			xerror.Panic(v.Init(cfg), "registry %s init error", k)
			// 注册中心只有一个, 所以可以使用Default, 否着需要使用Get()
			Default = v
		}
	}))
}
