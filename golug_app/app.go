package golug_app

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/golug/pkg/golug_utils"
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
	golug_env.Get(&Domain, "env_prefix", "golug", "golug_domain", "golug_prefix")
	if Domain = trim(lower(Domain)); Domain == "" {
		Domain = "golug"
		xlog.Warnf("[domain] prefix should be set, default: %s", Domain)
	}

	golug_env.Get(&Home, "home", "dir")

	// 使用前缀获取系统环境变量
	golug_env.Get(&Project, "project", "name", "server_name")

	golug_env.Get(&Mode, "mode", "run")

	Trace = golug_utils.IsTrue(trim(golug_env.GetEnv("trace")))
}

// CheckMod
// 运行环境检查
func CheckMod() error {
	var m = RunMode
	switch Mode {
	case m.Dev, m.Stag, m.Prod, m.Test, m.Release:
	default:
		return xerror.Fmt("running mode does not match, mode: %s", Mode)
	}

	return nil
}

func DefaultFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet("app", pflag.PanicOnError)
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
