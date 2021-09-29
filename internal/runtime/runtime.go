package runtime

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/healthy"
	"github.com/pubgo/lug/internal/cmds/protoc"
	"github.com/pubgo/lug/internal/cmds/restapi"
	"github.com/pubgo/lug/internal/cmds/swagger"
	v "github.com/pubgo/lug/internal/cmds/version"
	"github.com/pubgo/lug/logger"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/lug/runenv"
	"github.com/pubgo/lug/vars"
	"github.com/pubgo/lug/version"
	"github.com/pubgo/lug/watcher"
)

var logs *zap.Logger

func init() {
	logs = logger.On(func(log *zap.Logger) { logs = log.Named("runtime") })
}

var rootCmd = &cobra.Command{Use: runenv.Domain, Version: version.Version}

func init() {
	rootCmd.AddCommand(v.Cmd)
	rootCmd.AddCommand(healthy.Cmd)
	rootCmd.AddCommand(restapi.Cmd)
	rootCmd.AddCommand(protoc.Cmd)
	rootCmd.AddCommand(swagger.Cmd)
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
	logs.Sugar().Infof("signal [%s] trigger", runenv.Signal.String())
}

func start(ent entry.Runtime) (err error) {
	defer xerror.RespErr(&err)

	logs.Sugar().Info("before-start running")
	beforeList := append(entry.GetBeforeStartsList(), ent.Options().BeforeStarts...)
	for i := range beforeList {
		xerror.PanicF(xerror.Try(beforeList[i]), "before start error: %s", stack.Func(beforeList[i]))
	}
	logs.Sugar().Info("before-start over")

	xerror.Panic(ent.Start())

	logs.Sugar().Info("after-start running")
	afterList := append(entry.GetAfterStartsList(), ent.Options().AfterStarts...)
	for i := range afterList {
		xerror.PanicF(xerror.Try(afterList[i]), "after start error: %s", stack.Func(afterList[i]))
	}
	logs.Sugar().Info("after-start over")
	return
}

func stop(ent entry.Runtime) (err error) {
	defer xerror.RespErr(&err)

	logs.Sugar().Infof("service [%s] before-stop running", ent.Options().Name)
	beforeList := append(entry.GetBeforeStopsList(), ent.Options().BeforeStops...)
	for i := range beforeList {
		logger.Logs(beforeList[i], zap.String("msg", fmt.Sprintf("before stop error: %s", stack.Func(beforeList[i]))))
	}
	logs.Sugar().Infof("service [%s] before-stop over", ent.Options().Name)

	xerror.Panic(ent.Stop())

	logs.Sugar().Infof("service [%s] after-stop running", ent.Options().Name)
	afterList := append(entry.GetAfterStopsList(), ent.Options().AfterStops...)
	for i := range afterList {
		logger.Logs(afterList[i], zap.String("msg", fmt.Sprintf("after stop error: %s", stack.Func(afterList[i]))))
	}
	logs.Sugar().Infof("service [%s] after-stop over", ent.Options().Name)
	return nil
}

func Run(description string, entries ...entry.Entry) {
	defer xerror.RespExit()

	xerror.Assert(len(entries) == 0, "[entries] should not be zero")

	for _, ent := range entries {
		xerror.Assert(ent == nil, "[ent] should not be nil")

		_, ok := ent.(entry.Runtime)
		xerror.Assert(!ok, "[ent] not implement runtime")
	}

	rootCmd.Short = description
	rootCmd.Long = description
	rootCmd.PersistentFlags().AddFlagSet(runenv.DefaultFlags())
	rootCmd.PersistentFlags().AddFlagSet(config.DefaultFlags())
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error { return xerror.Wrap(cmd.Help()) }

	for i := range entries {
		ent := entries[i]
		entRT := ent.(entry.Runtime)
		cmd := entRT.Options().Command

		// 检查Command是否注册
		for _, c := range rootCmd.Commands() {
			xerror.Assert(c.Name() == cmd.Name(), "command(%s) already exists", cmd.Name())
		}

		// 注册plugin的command和flags
		// 先注册全局, 后注册项目相关
		xerror.TryThrow(func() {
			entPlugins := plugin.List(plugin.Module(entRT.Options().Name))
			for _, plg := range append(plugin.List(), entPlugins...) {
				cmd.PersistentFlags().AddFlagSet(plg.Flags())
				ent.Commands(plg.Commands())
				entRT.MiddlewareInter(plg.Middleware())
			}
		})

		cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) (err error) {
			defer xerror.RespErr(&err)

			// 项目名初始化
			runenv.Project = entRT.Options().Name

			xerror.TryThrow(func() {
				plugins := plugin.List(plugin.Module(runenv.Project))
				plugins = append(plugin.List(), plugins...)
				for _, plg := range plugins {

					// 注册watcher
					watcher.Watch(plg.String(), plg.Watch)

					// 注册debug
					healthy.Register(plg.String(), plg.Health())

					// 注册vars
					xerror.Panic(plg.Vars(vars.Watch))
				}
			})

			// config初始化
			xerror.Panic(config.Init())

			// watcher初始化
			xerror.Panic(watcher.Init())

			// plugin初始化
			xerror.TryThrow(func() {
				plugins := plugin.List(plugin.Module(runenv.Project))
				plugins = append(plugin.List(), plugins...)
				for _, plg := range plugins {
					xerror.PanicF(plg.Init(ent), "plugin [%s] init error", plg.String())
				}
			})

			// entry初始化
			xerror.PanicF(entRT.InitRT(), runenv.Project)
			return nil
		}

		cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
			defer xerror.RespErr(&err)
			xerror.Panic(start(entRT))
			handleSignal()
			xerror.Panic(stop(entRT))
			return nil
		}

		rootCmd.AddCommand(cmd)
	}

	xerror.Panic(rootCmd.Execute())
}

func Start(ent entry.Entry) {
	defer xerror.RespExit()

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

	// entry初始化
	xerror.Panic(entRun.InitRT())

	// plugin初始化
	plugins := plugin.List(plugin.Module(entRun.Options().Name))
	plugins = append(plugin.List(), plugins...)
	for _, pg := range plugins {
		key := pg.String()
		xerror.PanicF(pg.Init(ent), "plugin [%s] init error", key)

		// watch key
		watcher.Watch(key, pg.Watch)
	}

	xerror.Panic(start(entRun))
}

func Stop(ent entry.Entry) {
	defer xerror.RespExit()

	xerror.Assert(ent == nil, "[entry] should not be nil")

	entRun, ok := ent.(entry.Runtime)
	xerror.Assert(!ok, "[ent] not implement runtime")

	xerror.Panic(stop(entRun))
}
