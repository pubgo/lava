package golug_cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/pubgo/golug/golug_app"
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_plugin"
	"github.com/pubgo/golug/golug_watcher"
	"github.com/pubgo/golug/internal/golug_run"
	"github.com/pubgo/golug/version"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/pubgo/xprocess/xutil"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{Use: golug_app.Domain, Version: version.Version}

func initCfg() (err error) {
	defer xerror.RespErr(&err)

	// 处理所有的配置,环境变量和flag
	// 配置顺序, 默认值->环境变量->配置文件->flag->配置文件
	// 配置文件中可以设置环境变量
	// flag可以指定配置文件位置
	// 始化配置文件
	xerror.Panic(golug_config.Init())

	// 从配置文件加载
	if golug_config.CfgPath != "" {
		xerror.Panic(golug_config.InitWithCfgPath())
	}

	// 初始化框架, 加载环境变量, 加载本地配置
	// 初始化完毕所有的配置以及外部配置以及相关的参数和变量
	// 剩下的就是获取配置了
	if !golug_config.IsExist() {
		xerror.Panic(golug_config.InitWithDir())
	}

	xerror.Assert(golug_config.GetCfg().ConfigFileUsed() == "", "config file not found")

	xerror.ExitF(golug_config.GetCfg().ReadInConfig(), "read config failed")
	golug_config.InitHome()

	xerror.Panic(golug_config.InitApp())

	xerror.Panic(golug_app.CheckMod())

	xerror.Panic(golug_config.Fire())
	return nil
}

func handleSignal() {
	if golug_app.CatchSigpipe {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGPIPE)
		go func() {
			<-sigChan
			xlog.Warn("Caught SIGPIPE (ignoring all future SIGPIPEs)")
			signal.Ignore(syscall.SIGPIPE)
		}()
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGHUP)
	golug_app.Signal = <-ch
}

func start(ent golug_entry.RunEntry) (err error) {
	defer xerror.RespErr(&err)

	beforeStarts := golug_run.GetBeforeStarts()
	for i := range beforeStarts {
		func(i int) {
			defer xerror.RespRaise(func(err xerror.XErr) error {
				return err.WrapF("before start error: %s", xutil.FuncStack(beforeStarts[i]))
			})

			beforeStarts[i]()
		}(i)
	}

	xerror.Panic(ent.Start())

	afterStarts := golug_run.GetAfterStarts()
	for i := range afterStarts {
		func(i int) {
			defer xerror.RespRaise(func(err xerror.XErr) error {
				return err.WrapF("after start error: %s", xutil.FuncStack(afterStarts[i]))
			})

			afterStarts[i]()
		}(i)
	}

	return
}

func stop(ent golug_entry.RunEntry) (err error) {
	defer xerror.RespErr(&err)

	beforeStops := golug_run.GetBeforeStarts()
	for i := range beforeStops {
		func(i int) {
			defer xerror.RespRaise(func(err xerror.XErr) error {
				return err.WrapF("before stop error: %s", xutil.FuncStack(beforeStops[i]))
			})

			beforeStops[i]()
		}(i)
	}

	xerror.Panic(ent.Stop())

	afterStops := golug_run.GetAfterStops()
	for i := range afterStops {
		func(i int) {
			defer xerror.RespRaise(func(err xerror.XErr) error {
				return err.WrapF("after stop error: %s", xutil.FuncStack(afterStops[i]))
			})

			afterStops[i]()
		}(i)
	}

	return nil
}

func Run(entries ...golug_entry.Entry) (err error) {
	defer xerror.RespErr(&err)

	xerror.Assert(len(entries) == 0, "[entries] should not be zero")

	for _, ent := range entries {
		xerror.Assert(ent == nil, "[ent] should not be nil")

		entRun := ent.(golug_entry.RunEntry)
		opt := entRun.Options()
		xerror.Assert(opt.Name == "" || opt.Version == "", "[name,version] should not be empty")
	}

	rootCmd.PersistentFlags().AddFlagSet(golug_app.DefaultFlags())
	rootCmd.PersistentFlags().AddFlagSet(golug_config.DefaultFlags())
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error { return xerror.Wrap(cmd.Help()) }

	for _, ent := range entries {
		entRun := ent.(golug_entry.RunEntry)
		cmd := entRun.Options().Command

		// 检查Command是否注册
		for _, c := range rootCmd.Commands() {
			xerror.Assert(c.Name() == cmd.Name(), "command(%s) already exists", cmd.Name())
		}

		// 注册plugin的command和flags
		entPlugins := golug_plugin.List(golug_plugin.Module(entRun.Options().Name))
		for _, pl := range append(golug_plugin.List(), entPlugins...) {
			cmd.PersistentFlags().AddFlagSet(pl.Flags())
			ent.Commands(pl.Commands())
		}

		// 配置初始化
		cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) (err error) {
			defer xerror.RespErr(&err)
			golug_app.Project = entRun.Options().Name
			xerror.Panic(initCfg())
			return nil
		}

		cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
			defer xerror.RespErr(&err)

			// entry初始化
			xerror.Panic(entRun.Init())

			// 初始化组件, 初始化插件
			plugins := golug_plugin.List(golug_plugin.Module(entRun.Options().Name))
			plugins = append(golug_plugin.List(), plugins...)
			for _, pg := range plugins {
				key := pg.String()
				xerror.PanicF(pg.Init(ent), "plugin [%s] init error", key)

				// watch key
				golug_watcher.Watch(key, pg.Watch)
			}

			xerror.Panic(start(entRun))

			if golug_app.IsBlock {
				handleSignal()
			}

			xerror.Panic(stop(entRun))
			return nil
		}

		rootCmd.AddCommand(cmd)
	}

	return xerror.Wrap(rootCmd.Execute())
}
