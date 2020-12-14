package golug_env

import (
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/xerror"
	"github.com/spf13/pflag"
)

// 默认的全局配置
var (
	IsBlock           = true
	Domain            = "golug"
	Trace             = false
	Home              = filepath.Join(xerror.PanicStr(filepath.Abs(filepath.Dir(""))), "home")
	Project           = "golug"
	Mode              = "dev"
	Signal  os.Signal = syscall.Signal(0)
	// RunMode 项目运行模式
	RunMode = struct {
		Dev     string
		Test    string
		Stag    string
		Prod    string
		Release string
	}{
		Dev:     "dev",
		Test:    "test",
		Stag:    "stag",
		Prod:    "prod",
		Release: "release",
	}
)

func IsDev() bool {
	return Mode == RunMode.Dev
}

func IsTest() bool {
	return Mode == RunMode.Test
}

func IsStag() bool {
	return Mode == RunMode.Stag
}

func IsProd() bool {
	return Mode == RunMode.Prod
}

func IsRelease() bool {
	return Mode == RunMode.Release
}

func init() {
	// 从环境变量中获取系统默认值
	// 获取系统默认的前缀, 环境变量前缀等
	Get(&Domain, "env_prefix")
	Get(&Home, "home", "dir")

	// 使用前缀获取系统环境变量
	Get(&Project, "project", "name")
	Get(&Mode, "mode", "run")

	if trace := trim(GetEnv("trace")); trace != "" {
		Trace, _ = strconv.ParseBool(trace)
	}

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
