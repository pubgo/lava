package service

import (
	"fmt"
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/config/config_builder"
	"github.com/pubgo/lava/core/healthy"
	"github.com/pubgo/lava/core/logging/log_builder"
	"github.com/pubgo/lava/core/logging/logutil"
	"github.com/pubgo/lava/core/watcher"
	"os"
	"sort"

	"github.com/pubgo/lava/config/config_flag"
	"github.com/pubgo/lava/internal/envs"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/plugins/signal"
	"github.com/pubgo/lava/runtime"
	"github.com/pubgo/lava/service/service_type"
	"github.com/pubgo/lava/vars"
	"github.com/pubgo/lava/version"
	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

func Run(desc string, entries ...service_type.Service) {
	defer xerror.RespExit()

	xerror.Assert(len(entries) == 0, "[entries] should not be zero")

	for _, ent := range entries {
		xerror.Assert(ent == nil, "[ent] should not be nil")
	}

	var app = &cli.App{
		Name:    runtime.Domain,
		Version: version.Version,
	}

	app.Usage = desc
	app.Description = desc

	// 注册默认flags
	app.Flags = append(app.Flags, config_flag.Flags()...)

	// 注册全局plugin
	for _, plg := range plugin.All() {
		if plg == nil {
			continue
		}

		app.Flags = append(app.Flags, plg.Flags()...)

		var cmd = plg.Commands()
		if cmd != nil {
			// 检查Command是否注册
			xerror.Assert(app.Command(cmd.Name) != nil, "command(%s) already exists", cmd.Name)
			app.Commands = append(app.Commands, cmd)
		}

		// 注册健康检查
		if plg.Health() != nil {
			healthy.Register(plg.ID(), plg.Health())
		}

		// 注册vars
		xerror.Panic(plg.Vars(vars.Register))

		// 注册watcher
		watcher.Watch(plg.ID(), plg.Watch)
	}

	for i := range entries {
		ent := entries[i].(*serviceImpl)
		cmd := ent.command()

		// 检查项目Command是否注册
		xerror.Assert(app.Command(cmd.Name) != nil, "command(%s) already exists", cmd.Name)

		cmd.Action = func(ctx *cli.Context) error {
			// 项目名初始化
			runtime.Project = ent.Options().Name
			envs.SetName(version.Domain, runtime.Project)

			// 运行环境检查
			if _, ok := runtime.RunModeValue[runtime.Mode]; !ok {
				panic(fmt.Sprintf("mode(%s) not match in (%v)", runtime.Mode, runtime.RunModeValue))
			}

			// 本地配置初始化
			config_builder.Init()

			// 日志初始化
			log_builder.Init(config.GetCfg())

			// 插件初始化
			for _, plg := range append(plugin.All(), ent.plugins()...) {
				ent.middleware(plg.Middleware())
				ent.BeforeStarts(plg.BeforeStarts()...)
				ent.AfterStarts(plg.AfterStarts()...)
				ent.BeforeStops(plg.BeforeStops()...)
				ent.AfterStops(plg.AfterStops()...)

				logutil.LogOrPanic(zap.L(), fmt.Sprintf("plugin(%s) init", plg.ID()), func() error {
					return plg.Init(config.GetCfg())
				})
			}

			xerror.Panic(ent.init())
			xerror.Panic(ent.start())
			signal.Block()
			xerror.Panic(ent.stop())
			return nil
		}

		app.Commands = append(app.Commands, cmd)
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))
	xerror.Panic(app.Run(os.Args))
}
