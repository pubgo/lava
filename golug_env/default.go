package golug_env

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/xerror"
	"github.com/spf13/pflag"
)

// 默认的全局配置
var (
	IsBlock                 = true
	Domain                  = "golug"
	Trace                   = false
	Home                    = filepath.Join(xerror.PanicStr(filepath.Abs(filepath.Dir(""))), "home")
	Project                 = "golug"
	Mode                    = "dev"
	Signal        os.Signal = syscall.Signal(0)
	replacer                = strings.NewReplacer("-", "_", ".", "_", "/", "_")
	DefaultSecret           = "zpCjWPsbqK@@^hR01qLDmZcXhKRIZgjHfxSG2KA%J#bFp!7YQVSmzXGc!sE!^qSM7@&d%oXHQtpR7K*8eRTdhRKjaxF#t@bd#A!"
	RunMode                 = RunEnvMode{Dev: "dev", Test: "test", Stag: "stag", Prod: "prod", Release: "release"}
)

func init() {
	handleEnv()

	// 从环境变量中获取系统默认值
	// 获取系统默认的前缀, 环境变量前缀等
	GetSys(&Domain, "env_prefix")
	Get(&Home, "home", "dir")

	// 使用前缀获取系统环境变量
	Get(&Project, "project", "name")
	Get(&Mode, "mode", "run")
	Get(&DefaultSecret, "secret", "token", "app_secret", "app_token")

	Trace = IsTrue(trim(GetEnv("trace")))

	// 运行环境检查
	xerror.Exit(dix_run.WithBeforeStart(func(ctx *dix_run.BeforeStartCtx) {
		var m = RunMode
		switch Mode {
		case m.Dev, m.Stag, m.Prod, m.Test, m.Release:
		default:
			xerror.Panic(xerror.Fmt("running mode does not match, mode: %s", Mode))
		}
	}))
}

func DefaultFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet("app", pflag.PanicOnError)
	flags.StringVarP(&Mode, "mode", "m", Mode, "running mode(dev|test|stag|prod|release)")
	flags.BoolVarP(&Trace, "trace", "t", Trace, "enable trace")
	flags.StringVarP(&Project, "project", "p", Project, "project name")
	flags.BoolVarP(&IsBlock, "block", "b", IsBlock, "enable signal block")
	return flags
}

func handleEnv() {
	// 环境变量处理, 大写
	for _, env := range os.Environ() {
		if _envs := strings.SplitN(env, "=", 2); len(_envs) == 2 && trim(_envs[0]) != "" {
			_ = os.Unsetenv(_envs[0])
			key := replacer.Replace(upper(trim(_envs[0])))
			_ = os.Setenv(key, _envs[1])
		}
	}
}
