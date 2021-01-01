package golug_cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_app"
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_plugin"
	"github.com/pubgo/golug/golug_watcher"
	"github.com/pubgo/golug/version"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/spf13/cobra"
)

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

func Start(ent golug_entry.Entry) (err error) {
	defer xerror.RespErr(&err)

	// 配置顺序, 默认值->环境变量->配置文件->flag->配置文件
	// 配置文件中可以设置环境变量
	// flag可以指定配置文件位置
	// 始化配置文件
	xerror.Panic(golug_config.Init())

	// 初始化框架, 加载环境变量, 加载本地配置
	// 初始化完毕所有的配置以及外部配置以及相关的参数和变量
	// 剩下的就是获取配置了
	xerror.Panic(ent.Run().Init())

	// 初始化组件, 初始化插件
	plugins := golug_plugin.List(golug_plugin.Module(ent.Run().Options().Name))
	plugins = append(golug_plugin.List(), plugins...)
	for _, pg := range plugins {
		key := pg.String()
		xerror.PanicF(pg.Init(ent), "plugin [%s] init error", key)
		golug_watcher.Watch(key, pg.Watch)
	}

	xerror.Panic(dix_run.BeforeStart())
	xerror.Panic(ent.Run().Start())
	xerror.Panic(dix_run.AfterStart())

	return
}

func Stop(ent golug_entry.Entry) (err error) {
	defer xerror.RespErr(&err)

	xerror.Panic(dix_run.BeforeStop())
	xerror.Panic(ent.Run().Stop())
	xerror.Panic(dix_run.AfterStop())

	return nil
}

func Run(entries ...golug_entry.Entry) (err error) {
	defer xerror.RespErr(&err)

	if len(entries) == 0 {
		return xerror.New("[entries] should not be zero")
	}

	for _, ent := range entries {
		if ent == nil {
			return xerror.New("[ent] should not be nil")
		}

		opt := ent.Run().Options()
		if opt.Name == "" || opt.Version == "" {
			return xerror.New("neither [name] nor [version] can be empty")
		}
	}

	var rootCmd = &cobra.Command{Use: golug_app.Domain, Version: version.Version}
	rootCmd.PersistentFlags().AddFlagSet(golug_app.DefaultFlags())
	rootCmd.PersistentFlags().AddFlagSet(golug_config.DefaultFlags())
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error { return xerror.Wrap(cmd.Help()) }

	for _, ent := range entries {
		ent := ent
		cmd := ent.Run().Options().Command

		// 检查Command是否注册
		for _, c := range rootCmd.Commands() {
			if c.Name() == cmd.Name() {
				return xerror.Fmt("command(%s) already exists", cmd.Name())
			}
		}

		// 注册plugin的command和flags
		entPlugins := golug_plugin.List(golug_plugin.Module(ent.Run().Options().Name))
		for _, pl := range append(golug_plugin.List(), entPlugins...) {
			cmd.PersistentFlags().AddFlagSet(pl.Flags())
			ent.Commands(pl.Commands())
		}

		cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
			defer xerror.RespErr(&err)

			xerror.Panic(Start(ent))

			if golug_app.IsBlock {
				handleSignal()
			}

			xerror.Panic(Stop(ent))
			return nil
		}
		rootCmd.AddCommand(cmd)
	}

	return xerror.Wrap(rootCmd.Execute())
}
