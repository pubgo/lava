package golug

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/pubgo/golug/env"
	"github.com/pubgo/golug/gutils"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/spf13/pflag"
)

// 默认的全局配置
var (
	IsBlock                = true
	Domain                 = "golug"
	CatchSigpipe           = true
	Trace                  = false
	Home                   = filepath.Join(xerror.PanicStr(filepath.Abs(filepath.Dir(""))), "home")
	Project                = "golug"
	Mode                   = RunMode.Dev
	Signal       os.Signal = syscall.Signal(0)
	// RunEnvMode 项目运行模式
	RunMode = struct {
		Dev     string
		Test    string
		Stag    string
		Prod    string
		Release string
	}{Dev: "dev", Test: "test", Stag: "stag", Prod: "prod", Release: "release"}
	trim  = strings.TrimSpace
	lower = strings.ToLower
)

func init() {
	// 从环境变量中获取系统默认值
	// 获取系统默认的前缀, 环境变量前缀等
	env.GetVal(&Domain, "env_prefix", "golug", "golug_domain", "golug_prefix")
	if Domain = trim(lower(Domain)); Domain == "" {
		Domain = "golug"
		xlog.Warnf("[domain] prefix should be set, default: %s", Domain)
	}

	env.GetVal(&Home, "home", "dir")

	// 使用前缀获取系统环境变量
	env.GetVal(&Project, "project", "name", "server_name")

	env.GetVal(&Mode, "mode", "run")

	Trace = gutils.IsTrue(trim(env.Get("trace")))
}

func DefaultFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet("app", pflag.PanicOnError)
	flags.StringVarP(&Home, "home", "c", Mode, "config home")
	flags.StringVarP(&Mode, "mode", "m", Mode, "running mode(dev|test|stag|prod|release)")
	flags.BoolVarP(&Trace, "trace", "t", Trace, "enable trace")
	flags.StringVarP(&Project, "project", "p", Project, "project name")
	flags.BoolVarP(&IsBlock, "block", "b", IsBlock, "enable signal block")
	flags.BoolVar(&CatchSigpipe, "catch-sigpipe", CatchSigpipe, "catch and ignore SIGPIPE on stdout and stderr if specified")
	return flags
}

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
