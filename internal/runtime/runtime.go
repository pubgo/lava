package runtime

import (
	"os"
	"os/signal"
	"sort"
	"syscall"

	"github.com/pubgo/dix"
	"github.com/pubgo/x/q"
	"github.com/pubgo/x/stack"
	"github.com/pubgo/x/strutil"
	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/entry"
	"github.com/pubgo/lava/healthy"
	"github.com/pubgo/lava/internal/cmds/restapi"
	v "github.com/pubgo/lava/internal/cmds/version"
	"github.com/pubgo/lava/internal/logz"
	"github.com/pubgo/lava/logger"
	"github.com/pubgo/lava/pkg/syncx"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/plugins/watcher"
	"github.com/pubgo/lava/runenv"
	"github.com/pubgo/lava/vars"
	"github.com/pubgo/lava/version"
)

const name = "runtime"

var logs = logz.New(name)
var app = &cli.App{
	Name:    runenv.Domain,
	Version: version.Version,
	Commands: cli.Commands{
		v.Cmd,
		healthy.Cmd,
		restapi.Cmd,
	},
}

func handleSignal() {
	if runenv.CatchSigpipe {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGPIPE)
		syncx.GoSafe(func() {
			<-sigChan
			logs.Warn("Caught SIGPIPE (ignoring all future SIGPIPE)")
			signal.Ignore(syscall.SIGPIPE)
		})
	}

	if !runenv.Block {
		return
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGHUP)
	runenv.Signal = <-ch
	logs.Infof("signal [%s] trigger", runenv.Signal.String())
}

func start(ent entry.Runtime) {
	logs.StepAndThrow("before-start running", func() error {
		beforeList := append(entry.GetBeforeStartsList(), ent.Options().BeforeStarts...)
		for i := range beforeList {
			logs.Infof("running %s", stack.Func(beforeList[i]))
			xerror.PanicF(xerror.Try(beforeList[i]), stack.Func(beforeList[i]))
		}
		return nil
	})

	logs.StepAndThrow("server start", ent.Start)

	logs.StepAndThrow("after-start running", func() error {
		afterList := append(entry.GetAfterStartsList(), ent.Options().AfterStarts...)
		for i := range afterList {
			logs.Infof("running %s", stack.Func(afterList[i]))
			xerror.PanicF(xerror.Try(afterList[i]), stack.Func(afterList[i]))
		}
		return nil
	})
}

func stop(ent entry.Runtime) {
	logs.Step("before-stop running", func() error {
		beforeList := append(entry.GetBeforeStopsList(), ent.Options().BeforeStops...)
		for i := range beforeList {
			logs.Infof("running %s", stack.Func(beforeList[i]))
			xerror.PanicF(xerror.Try(beforeList[i]), stack.Func(beforeList[i]))
		}
		return nil
	})

	logs.Step("server stop", ent.Stop)

	logs.Step("after-stop running", func() error {
		afterList := append(entry.GetAfterStopsList(), ent.Options().AfterStops...)
		for i := range afterList {
			logs.Infof("running %s", stack.Func(afterList[i]))
			xerror.PanicF(xerror.Try(afterList[i]), stack.Func(afterList[i]))
		}
		return nil
	})
}

func Run(description string, entries ...entry.Entry) {
	defer xerror.RespExit()

	xerror.Assert(len(entries) == 0, "[entries] should not be zero")

	for _, ent := range entries {
		xerror.Assert(ent == nil, "[ent] should not be nil")

		_, ok := ent.(entry.Runtime)
		xerror.Assert(!ok, "[ent] not implement runtime, \n%s", q.Sq(ent))
	}

	app.Usage = description
	app.Description = description
	app.Flags = append(app.Flags, config.DefaultFlags()...)

	for i := range entries {
		ent := entries[i]
		entRT := ent.(entry.Runtime)
		cmd := entRT.Options().Command

		// 检查Command是否注册
		xerror.Assert(app.Command(cmd.Name) != nil, "command(%s) already exists", cmd.Name)

		// 注册plugin的command和flags
		// 先注册全局, 后注册项目相关
		for _, plg := range plugin.All() {
			cmd.Flags = append(cmd.Flags, plg.Flags()...)
			ent.Commands(plg.Commands())
			entRT.MiddlewareInter(plg.Middleware())
		}

		cmd.Before = func(ctx *cli.Context) error {
			defer xerror.RespExit()
			var exclude = func() []string {
				if entRT.Options().Exclude == nil {
					return nil
				}
				return entRT.Options().Exclude()
			}()

			// 项目名初始化
			runenv.Project = entRT.Options().Name

			// config初始化
			xerror.Panic(config.Init())

			// 配置依赖注入
			xerror.Exit(dix.Provider(config.GetCfg()))

			// plugin初始化
			for _, plg := range plugin.All() {
				if strutil.Contain(exclude, plg.UniqueName()) {
					continue
				}

				// 注册watcher
				if plg.Watch() != nil {
					logs.LogAndThrow("plugin register watcher",
						func() error { watcher.Watch(plg.UniqueName(), plg.Watch()); return nil },
						logger.Name(plg.UniqueName()),
					)
				}

				// 注册健康检查
				healthy.Register(plg.UniqueName(), plg.Health())

				// 注册vars
				xerror.Panic(plg.Vars(vars.Watch))

				logs.Logs("plugin init", plg.Init, logger.Name(plg.UniqueName()))
			}

			// entry初始化
			entRT.InitRT()

			// watcher初始化, 最后初始化, 从远程获取最新的配置
			xerror.Panic(watcher.Init())
			return nil
		}

		cmd.Action = func(ctx *cli.Context) error {
			defer xerror.RespExit()
			start(entRT)
			handleSignal()
			stop(entRT)
			return nil
		}

		app.Commands = append(app.Commands, cmd)
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))
	xerror.Panic(app.Run(os.Args))
}
