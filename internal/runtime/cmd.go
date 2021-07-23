package runtime

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pubgo/dix"
	"github.com/pubgo/x/stack"
	"github.com/pubgo/x/try"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/healthy"
	v "github.com/pubgo/lug/internal/cmds/version"
	"github.com/pubgo/lug/logutil"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/lug/runenv"
	"github.com/pubgo/lug/version"
	"github.com/pubgo/lug/watcher"
)

var logs = xlog.GetLogger("runtime")
var rootCmd = &cobra.Command{Use: runenv.Domain, Version: version.Version}

func init() {
	rootCmd.AddCommand(v.Cmd)
	rootCmd.AddCommand(healthy.Cmd)
}

func handleSignal() {
	if runenv.CatchSigpipe {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGPIPE)
		go func() {
			<-sigChan
			logs.Warn("Caught SIGPIPE (ignoring all future SIGPIPEs)")
			signal.Ignore(syscall.SIGPIPE)
		}()
	}

	if !runenv.Block {
		return
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGHUP)
	runenv.Signal = <-ch
	logs.Infof("signal [%s] trigger", runenv.Signal.String())
}

func start(ent entry.Runtime, args []string) (err error) {
	defer xerror.RespErr(&err)

	logs.Infof("service [%s] before-start running", ent.Options().Name)
	bStarts := append(entry.GetBeforeStartsList(), ent.Options().BeforeStarts...)
	for i := range bStarts {
		xerror.PanicF(try.Try(bStarts[i]), "before start error: %s", stack.Func(bStarts[i]))
	}
	logs.Infof("service [%s] before-start over", ent.Options().Name)

	xerror.Panic(ent.Start())

	logs.Infof("service [%s] after-start running", ent.Options().Name)
	aStarts := append(entry.GetAfterStartsList(), ent.Options().AfterStarts...)
	for i := range aStarts {
		xerror.PanicF(try.Try(aStarts[i]), "after start error: %s", stack.Func(bStarts[i]))
	}
	logs.Infof("service [%s] after-start over", ent.Options().Name)
	return
}

func stop(ent entry.Runtime) (err error) {
	defer xerror.RespErr(&err)

	logs.Infof("service [%s] before-stop running", ent.Options().Name)
	bStops := append(entry.GetBeforeStopsList(), ent.Options().BeforeStops...)
	for i := range bStops {
		logutil.Logs(bStops[i], zap.String("msg", fmt.Sprintf("before stop error: %s", stack.Func(bStops[i]))))
	}
	logs.Infof("service [%s] before-stop over", ent.Options().Name)

	xerror.Panic(ent.Stop())

	logs.Infof("service [%s] after-stop running", ent.Options().Name)
	aStops := append(entry.GetAfterStopsList(), ent.Options().AfterStops...)
	for i := range aStops {
		logutil.Logs(aStops[i], zap.String("msg", fmt.Sprintf("after stop error: %s", stack.Func(aStops[i]))))
	}
	logs.Infof("service [%s] after-stop over", ent.Options().Name)
	return nil
}

func Run(short string, entries ...entry.Entry) (err error) {
	defer xerror.RespErr(&err)

	xerror.Assert(len(entries) == 0, "[entries] should not be zero")

	for _, ent := range entries {
		xerror.Assert(ent == nil, "[ent] should not be nil")

		_, ok := ent.(entry.Runtime)
		xerror.Assert(!ok, "[ent] not implement runtime")
	}

	rootCmd.Short = short
	rootCmd.Long = short
	rootCmd.PersistentFlags().AddFlagSet(runenv.DefaultFlags())
	rootCmd.PersistentFlags().AddFlagSet(config.DefaultFlags())
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error { return xerror.Wrap(cmd.Help()) }

	for i := range entries {
		ent := entries[i]
		entRun := ent.(entry.Runtime)
		cmd := entRun.Options().Command

		// 检查Command是否注册
		for _, c := range rootCmd.Commands() {
			xerror.Assert(c.Name() == cmd.Name(), "command(%s) already exists", cmd.Name())
		}

		// 注册plugin的command和flags
		entPlugins := plugin.List(plugin.Module(entRun.Options().Name))
		for _, pl := range append(plugin.List(), entPlugins...) {
			cmd.PersistentFlags().AddFlagSet(pl.Flags())
			ent.Commands(pl.Commands())
		}

		// 配置初始化
		cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) (err error) {
			defer xerror.RespErr(&err)

			// 本项目名字初始化
			runenv.Project = entRun.Options().Name

			// 配置初始化
			xerror.Panic(config.Init())
			xerror.Panic(dix.Dix(config.GetCfg()))

			// 初始化组件, 初始化插件
			plugins := plugin.List(plugin.Module(runenv.Project))
			plugins = append(plugin.List(), plugins...)
			for _, pg := range plugins {
				key := pg.String()
				xerror.PanicF(pg.Init(ent), "plugin [%s] init error", key)

				// watch初始化, watch remote key
				watcher.Watch(key, pg)
			}

			xerror.Panic(watcher.Init())

			// entry初始化
			xerror.PanicF(entRun.InitRT(), runenv.Project)
			return nil
		}

		cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
			defer xerror.RespErr(&err)

			xerror.Panic(start(entRun, args))
			handleSignal()
			xerror.Panic(stop(entRun))
			return nil
		}

		rootCmd.AddCommand(cmd)
	}

	return xerror.Wrap(rootCmd.Execute())
}

func Start(ent entry.Entry, args ...string) (err error) {
	defer xerror.RespErr(&err)

	xerror.Assert(ent == nil, "[entry] should not be nil")

	entRun, ok := ent.(entry.Runtime)
	xerror.Assert(!ok, "[ent] not implement runtime")

	opt := entRun.Options()
	xerror.Assert(opt.Name == "", "[name] should not be empty")

	entPlugins := plugin.List(plugin.Module(entRun.Options().Name))
	for _, pl := range append(plugin.List(), entPlugins...) {
		// 加载flag
		_ = pl.Flags()
	}

	// config初始化
	runenv.Project = entRun.Options().Name
	xerror.Panic(config.Init())
	xerror.Panic(watcher.Init())
	xerror.Panic(dix.Dix(config.GetCfg()))

	// entry初始化
	xerror.Panic(entRun.InitRT())

	// plugin初始化
	plugins := plugin.List(plugin.Module(entRun.Options().Name))
	plugins = append(plugin.List(), plugins...)
	for _, pg := range plugins {
		key := pg.String()
		xerror.PanicF(pg.Init(ent), "plugin [%s] init error", key)

		// watch key
		watcher.Watch(key, pg)
	}

	return xerror.Wrap(start(entRun, args))
}

func Stop(ent entry.Entry) (err error) {
	defer xerror.RespErr(&err)

	xerror.Assert(ent == nil, "[entry] should not be nil")

	entRun, ok := ent.(entry.Runtime)
	xerror.Assert(!ok, "[ent] not implement runtime")

	return xerror.Wrap(stop(entRun))
}
