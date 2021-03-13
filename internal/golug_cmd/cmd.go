package golug_cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/pubgo/golug/config"
	"github.com/pubgo/golug/entry"
	"github.com/pubgo/golug/internal/golug_run"
	"github.com/pubgo/golug/plugin"
	"github.com/pubgo/golug/version"
	"github.com/pubgo/golug/watcher"
	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{Use: config.Domain, Version: version.Version}

func handleSignal() {
	if config.CatchSigpipe {
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
	config.Signal = <-ch
}

func start(ent entry.RunEntry) (err error) {
	defer xerror.RespErr(&err)

	beforeStarts := golug_run.GetBeforeStarts()
	beforeStarts = append(beforeStarts, ent.Options().BeforeStarts...)
	for i := range beforeStarts {
		func(i int) {
			defer xerror.RespRaise(func(err xerror.XErr) error {
				return err.WrapF("before start error: %s", stack.Func(beforeStarts[i]))
			})

			beforeStarts[i]()
		}(i)
	}

	xerror.Panic(ent.Start())

	afterStarts := golug_run.GetAfterStarts()
	afterStarts = append(afterStarts, ent.Options().AfterStarts...)
	for i := range afterStarts {
		func(i int) {
			defer xerror.RespRaise(func(err xerror.XErr) error {
				return err.WrapF("after start error: %s", stack.Func(afterStarts[i]))
			})

			afterStarts[i]()
		}(i)
	}

	return
}

func stop(ent entry.RunEntry) (err error) {
	defer xerror.RespErr(&err)

	beforeStops := golug_run.GetBeforeStops()
	beforeStops = append(beforeStops, ent.Options().BeforeStops...)
	for i := range beforeStops {
		func(i int) {
			defer xerror.RespRaise(func(err xerror.XErr) error {
				return err.WrapF("before stop error: %s", stack.Func(beforeStops[i]))
			})

			beforeStops[i]()
		}(i)
	}

	xerror.Panic(ent.Stop())

	afterStops := golug_run.GetAfterStops()
	afterStops = append(afterStops, ent.Options().AfterStops...)
	for i := range afterStops {
		func(i int) {
			defer xerror.RespRaise(func(err xerror.XErr) error {
				return err.WrapF("after stop error: %s", stack.Func(afterStops[i]))
			})

			afterStops[i]()
		}(i)
	}

	return nil
}

func Run(entries ...entry.Entry) (err error) {
	defer xerror.RespErr(&err)

	xerror.Assert(len(entries) == 0, "[entries] should not be zero")

	for _, ent := range entries {
		xerror.Assert(ent == nil, "[ent] should not be nil")

		entRun := ent.(entry.RunEntry)
		opt := entRun.Options()
		xerror.Assert(opt.Name == "" || opt.Version == "", "[name,version] should not be empty")
	}

	rootCmd.PersistentFlags().AddFlagSet(config.DefaultFlags())
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error { return xerror.Wrap(cmd.Help()) }

	for _, ent := range entries {
		entRun := ent.(entry.RunEntry)
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
			config.Project = entRun.Options().Name
			xerror.Panic(config.Init())
			return nil
		}

		cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
			defer xerror.RespErr(&err)

			// entry初始化
			xerror.Panic(entRun.Init())

			// 初始化组件, 初始化插件
			plugins := plugin.List(plugin.Module(entRun.Options().Name))
			plugins = append(plugin.List(), plugins...)
			for _, pg := range plugins {
				key := pg.String()
				xerror.PanicF(pg.Init(ent), "plugin [%s] init error", key)

				// watch key
				watcher.Watch(key, pg.Watch)
			}

			xerror.Panic(start(entRun))

			if config.IsBlock {
				handleSignal()
			}

			xerror.Panic(stop(entRun))
			return nil
		}

		rootCmd.AddCommand(cmd)
	}

	return xerror.Wrap(rootCmd.Execute())
}
