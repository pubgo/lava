package registry

import (
	"github.com/pubgo/golug/config"
	"github.com/pubgo/golug/internal/golug_run"
	"github.com/pubgo/golug/gutils"
	"github.com/pubgo/xerror"
)

func init() {
	// 配置初始化完毕后, 解析registry配置
	config.On(func(_ *config.Config) {
		gutils.Mergo(&cfg, GetDefaultCfg())
		config.Decode(Name, &cfg)
	})

	golug_run.BeforeStart(func() {
		for k, v := range List() {
			xerror.PanicF(v.Init(cfg), "registry %s init error", k)
			// 注册中心只有一个, 所以可以使用Default, 否着需要使用Get()
			Default = v
		}
	})
}
