package runtime

import (
	"fmt"
	"os"
	"sort"

	"github.com/pubgo/x/q"
	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/entry"
	"github.com/pubgo/lava/internal/envs"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/log_config"
	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/plugins/healthy"
	"github.com/pubgo/lava/plugins/signal"
	"github.com/pubgo/lava/runtime"
	"github.com/pubgo/lava/vars"
	"github.com/pubgo/lava/version"
	"github.com/pubgo/lava/watcher"
)

var logs = logging.Component("runtime")
var app = &cli.App{
	Name:    runtime.Domain,
	Version: version.Version,
}

func Run(desc string, entries ...entry.Entry) {
	defer xerror.RespExit()

	xerror.Assert(len(entries) == 0, "[entries] should not be zero")

	for _, ent := range entries {
		xerror.Assert(ent == nil, "[ent] should not be nil")

		_, ok := ent.(entry.Runtime)
		xerror.Assert(!ok, "[ent] not implement runtime, \n%s", q.Sq(ent))
	}

	app.Usage = desc
	app.Description = desc

	// 注册默认flags
	app.Flags = append(app.Flags, config.DefaultFlags()...)

	// 注册全局plugin
	for _, plg := range plugin.All() {
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
		ent := entries[i]
		entRT := ent.(entry.Runtime)
		cmd := entRT.Options().Command

		// 检查项目Command是否注册
		xerror.Assert(app.Command(cmd.Name) != nil, "command(%s) already exists", cmd.Name)

		cmd.Action = func(ctx *cli.Context) error {
			// 项目名初始化
			runtime.Project = entRT.Options().Name
			envs.SetName(version.Domain, runtime.Project)

			// 运行环境检查
			if _, ok := runtime.RunModeValue[runtime.Mode]; !ok {
				panic(fmt.Sprintf("mode(%s) not match in (%v)", runtime.Mode, runtime.RunModeValue))
			}

			// 本地配置初始化
			config.Init()
			for _, plg := range plugin.All() {
				plg.InitCfg(config.GetCfg())
			}

			// 日志初始化
			logging.Init(func(cfg *log_config.Config) {
				xerror.Panic(config.GetMap(logging.Name).Decode(cfg))
			})

			// 插件初始化
			for _, plg := range plugin.All() {
				entRT.MiddlewareInter(plg.Middleware())
				logutil.OkOrPanic(logs.L(), "plugin init", plg.Init, zap.String("plugin-name", plg.ID()))
			}

			// entry初始化, 项目初始化
			entRT.InitRT()

			start(entRT)
			signal.Block()
			stop(entRT)
			return nil
		}

		app.Commands = append(app.Commands, cmd)
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))
	xerror.Panic(app.Run(os.Args))
}

func start(ent entry.Runtime) {
	logutil.OkOrPanic(logs.L(), "before-start running", func() error {
		var beforeList []func()
		for _, p := range plugin.All() {
			beforeList = append(beforeList, p.BeforeStarts()...)
		}
		beforeList = append(beforeList, ent.Options().BeforeStarts...)
		for i := range beforeList {
			logs.S().Infof("running %s", stack.Func(beforeList[i]))
			xerror.PanicF(xerror.Try(beforeList[i]), stack.Func(beforeList[i]))
		}
		return nil
	})
	logutil.OkOrPanic(logs.L(), "server start", ent.Start)
	logutil.OkOrPanic(logs.L(), "after-start running", func() error {
		var afterList []func()
		for _, p := range plugin.All() {
			afterList = append(afterList, p.AfterStarts()...)
		}
		afterList = append(afterList, ent.Options().AfterStarts...)
		for i := range afterList {
			logs.S().Infof("running %s", stack.Func(afterList[i]))
			xerror.PanicF(xerror.Try(afterList[i]), stack.Func(afterList[i]))
		}
		return nil
	})
}

func stop(ent entry.Runtime) {
	logutil.OkOrErr(logs.L(), "before-stop running", func() error {
		var beforeList []func()
		for _, p := range plugin.All() {
			beforeList = append(beforeList, p.BeforeStops()...)
		}
		beforeList = append(beforeList, ent.Options().BeforeStops...)
		for i := range beforeList {
			logs.S().Infof("running %s", stack.Func(beforeList[i]))
			xerror.PanicF(xerror.Try(beforeList[i]), stack.Func(beforeList[i]))
		}
		return nil
	})

	logutil.OkOrErr(logs.L(), "server stop", ent.Stop)

	logutil.OkOrErr(logs.L(), "after-stop running", func() error {
		var afterList []func()
		for _, p := range plugin.All() {
			afterList = append(afterList, p.AfterStops()...)
		}
		afterList = append(afterList, ent.Options().AfterStops...)
		for i := range afterList {
			logs.S().Infof("running %s", stack.Func(afterList[i]))
			xerror.PanicF(xerror.Try(afterList[i]), stack.Func(afterList[i]))
		}
		return nil
	})
}
