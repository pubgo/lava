package app

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/healthy"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/plugins/watcher"
	"github.com/pubgo/lava/runenv"
	"github.com/pubgo/lava/vars"
)

func Init(name string) {
	defer xerror.RespExit()

	runenv.Project = name

	// 配置初始化
	xerror.Panic(config.Init())

	// 配置依赖注入
	xerror.Exit(dix.Provider(config.GetCfg()))

	// 获取本项目所有plugin
	plugins := plugin.All()
	for _, plg := range plugins {

		// 注册watcher
		watcher.Watch(plg.UniqueName(), plg.Watch())

		// 注册健康检查
		healthy.Register(plg.UniqueName(), plg.Health())

		// 注册vars
		xerror.Panic(plg.Vars(vars.Watch))

		// plugin初始化
		xerror.PanicF(plg.Init(), "plugin [%s] init error", plg.UniqueName())
	}
}
